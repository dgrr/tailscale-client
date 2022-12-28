package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dgrr/tl"
	"github.com/gen2brain/beeep"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.design/x/clipboard"
	"tailscale.com/client/tailscale"
	"tailscale.com/client/tailscale/apitype"
	"tailscale.com/ipn"
	"tailscale.com/ipn/ipnstate"
	"tailscale.com/types/key"
)

// App struct
type App struct {
	ctx           context.Context
	client        tailscale.LocalClient
	fileMod       chan struct{}
	initClipboard sync.Once
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

func notify(msg string) {
	beeep.Notify("Tailscale", msg, "icon/on.png")
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (app *App) startup(ctx context.Context) {
	app.ctx = ctx
	app.fileMod = make(chan struct{}, 1)

	notify("Tailscale started")

	go app.watchFiles()
	go app.watchIPN()

	runtime.EventsOn(app.ctx, "file_upload", func(data ...interface{}) {
		fmt.Println(data)
	})
}

func (app *App) watchFiles() {
	prevFiles := 0
	for {
		select {
		case <-time.After(time.Second * 10):
		case <-app.fileMod:
		}

		files, err := app.client.AwaitWaitingFiles(app.ctx, time.Second)
		if err != nil {
			log.Println(err)
		}

		if len(files) != prevFiles {
			prevFiles = len(files)
			runtime.EventsEmit(app.ctx, "update_files")
		}
	}
}

func (app *App) watchIPN() {
	for {
		watcher, err := app.client.WatchIPNBus(app.ctx, 0)
		if err != nil {
			log.Printf("loading IPN bus watcher: %s\n", err)
			time.Sleep(time.Second)
			continue
		}

		for {
			not, err := watcher.Next()
			if err != nil {
				log.Printf("Watching IPN Bus: %s\n", err)
				break
			}

			if not.FilesWaiting != nil {
				app.fileMod <- struct{}{}
			}

			runtime.EventsEmit(app.ctx, "update_all")

			log.Printf("IPN bus update: %v\n", not)
		}
	}
}

func (app *App) UploadFile(dnsName string) {
	status, err := app.client.Status(app.ctx)
	if err != nil {
		panic(err)
	}

	peers := status.Peers()

	i := tl.SearchFn(peers, func(nodeKey key.NodePublic) bool {
		peer := status.Peer[nodeKey]
		return peer.DNSName == dnsName
	})
	if i == -1 {
		return
	}

	peer := status.Peer[peers[i]]

	filename, err := runtime.OpenFileDialog(app.ctx, runtime.OpenDialogOptions{
		DefaultDirectory: func() string {
			dir, _ := os.UserHomeDir()
			return dir
		}(),
	})
	if err != nil {
		panic(err)
	}

	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	stat, _ := file.Stat()

	err = app.client.PushFile(app.ctx, peer.ID, stat.Size(), stat.Name(), file)
	if err != nil {
		log.Printf("error uploading file to %s: %s\n", dnsName, err)
	}
}

func (app *App) AcceptFile(filename string) {
	dir, err := runtime.OpenDirectoryDialog(app.ctx, runtime.OpenDialogOptions{
		DefaultDirectory: func() string {
			dir, _ := os.UserHomeDir()
			return dir
		}(),
	})
	if err != nil {
		panic(err)
	}
	defer func() {
		app.RemoveFile(filename)
	}()

	r, _, err := app.client.GetWaitingFile(app.ctx, filename)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	file, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, _ = io.Copy(file, r)
}

func (app *App) RemoveFile(filename string) {
	log.Printf("Removing file %s\n", filename)

	err := app.client.DeleteWaitingFile(app.ctx, filename)
	if err != nil {
		log.Printf("Removing file: %s: %s\n", filename, err)
	}

	app.fileMod <- struct{}{}
}

func (app *App) CurrentAccount() string {
	current, _, err := app.client.ProfileStatus(app.ctx)
	if err != nil {
		panic(err)
	}

	return current.Name
}

func (app *App) SetExitNode(dnsName string) {
	status, err := app.client.Status(app.ctx)
	if err != nil {
		panic(err)
	}

	peers := status.Peers()

	i := tl.SearchFn(peers, func(nodeKey key.NodePublic) bool {
		peer := status.Peer[nodeKey]
		return peer.DNSName == dnsName
	})
	if i == -1 {
		return
	}

	peer := status.Peer[peers[i]]

	prefs := &ipn.MaskedPrefs{
		Prefs:         ipn.Prefs{},
		ExitNodeIPSet: true,
		ExitNodeIDSet: true,
	}

	if !peer.ExitNode {
		err = prefs.SetExitNodeIP(peer.DNSName, status)
		if err != nil {
			panic(err)
		}
	}

	_, err = app.client.EditPrefs(app.ctx, prefs)
	if err != nil {
		log.Println(err)
	}

	runtime.EventsEmit(app.ctx, "exit_node_connect")

	if peer.ExitNode {
		notify(fmt.Sprintf("Removed exit node %s", peer.DNSName))
	} else {
		notify(fmt.Sprintf("Using %s as exit node", peer.DNSName))
	}
}

func (app *App) CopyClipboard(s string) {
	app.initClipboard.Do(func() {
		if err := clipboard.Init(); err != nil {
			panic(err)
		}
	})
	log.Printf("Copying \"%s\" to the clipboard\n", s)
	clipboard.Write(clipboard.FmtText, []byte(s))
}

func (app *App) Accounts() []string {
	current, all, err := app.client.ProfileStatus(app.ctx)
	if err != nil {
		panic(err)
	}

	names := tl.Filter(
		tl.Map(all, func(profile ipn.LoginProfile) string {
			return profile.Name
		}),
		func(name string) bool {
			return name != current.Name
		},
	)

	return names
}

type Instance struct {
	Files      []File      `json:"files"`
	Namespaces []Namespace `json:"namespaces"`
}

type Namespace struct {
	Name  string `json:"name"`
	Peers []Peer `json:"peers"`
}

type Peer struct {
	DNSName        string    `json:"dns_name"`
	Name           string    `json:"name"`
	ExitNode       bool      `json:"exit_node"`
	ExitNodeOption bool      `json:"exit_node_option"`
	Online         bool      `json:"online"`
	OS             string    `json:"os"`
	Addrs          []string  `json:"addrs"`
	Routes         []string  `json:"routes"` // primary routes
	IPs            []string  `json:"ips"`
	Created        time.Time `json:"created_at"`
	LastSeen       time.Time `json:"last_seen"`
}

type File struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

func (app *App) Self() Peer {
	log.Printf("Requesting self")

	status, err := app.client.Status(app.ctx)
	if err != nil {
		log.Printf("Requesting self: %s\n", err)
		return Peer{}
	}

	self := status.Self
	return convertPeer(self)
}

func convertPeer(status *ipnstate.PeerStatus) Peer {
	peerName, _ := splitPeerNamespace(status.DNSName)
	return Peer{
		DNSName:        status.DNSName,
		Name:           peerName,
		ExitNode:       status.ExitNode,
		ExitNodeOption: status.ExitNodeOption,
		Online:         status.Online,
		OS:             status.OS,
		Addrs:          status.Addrs,
		Created:        status.Created,
		LastSeen:       status.LastSeen,
		Routes: func() []string {
			if status.PrimaryRoutes == nil {
				return nil
			}

			return tl.Map(status.PrimaryRoutes.AsSlice(), func(prefix netip.Prefix) string {
				return prefix.String()
			})
		}(),
		IPs: tl.Map(status.TailscaleIPs, func(ip netip.Addr) string {
			return ip.String()
		}),
	}
}

func splitPeerNamespace(dnsName string) (peerName, namespace string) {
	names := strings.Split(dnsName, ".")
	namespace = strings.Join(names[1:], ".")
	peerName = names[0]
	return peerName, namespace
}

func (app *App) Files() []File {
	files, err := app.client.AwaitWaitingFiles(app.ctx, time.Second)
	if err != nil {
		log.Println(err)
		return nil
	}

	return tl.Map(files, func(file apitype.WaitingFile) File {
		return File{
			Name: file.Name,
			Size: file.Size,
		}
	})
}

func (app *App) Namespaces() []Namespace {
	status, err := app.client.Status(app.ctx)
	if err != nil {
		log.Printf("requesting instance: %s\n", err)
		return nil
	}

	res := make([]Namespace, 0)

	for _, nodeKey := range status.Peers() {
		tsPeer := status.Peer[nodeKey]
		_, namespace := splitPeerNamespace(tsPeer.DNSName)

		peer := convertPeer(tsPeer)

		i := tl.SearchFn(res, func(a Namespace) bool {
			return namespace == a.Name
		})
		if i == -1 {
			res = append(res, Namespace{
				Name: namespace,
				Peers: []Peer{
					peer,
				},
			})
		} else {
			res[i].Peers = append(res[i].Peers, peer)
		}
	}

	return res
}

func (app *App) SwitchTo(account string) {
	current, all, err := app.client.ProfileStatus(app.ctx)
	if err != nil {
		panic(err)
	}

	if account == current.Name {
		return
	}

	all = tl.Filter(all, func(profile ipn.LoginProfile) bool {
		return profile.Name == account
	})
	if len(all) == 0 {
		log.Printf("Profile %s not found\n", account)
		return
	}

	app.client.SwitchProfile(app.ctx, all[0].ID)
}
