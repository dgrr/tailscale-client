package main

import (
	"context"
	"log"
	"net/netip"
	"strings"
	"time"

	"github.com/dgrr/tl"
	"github.com/gen2brain/beeep"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"tailscale.com/client/tailscale"
	"tailscale.com/ipn"
	"tailscale.com/ipn/ipnstate"
	"tailscale.com/types/key"
)

// App struct
type App struct {
	ctx    context.Context
	client tailscale.LocalClient
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (app *App) startup(ctx context.Context) {
	app.ctx = ctx
	beeep.Notify("Tailscale", "tailscale started", "icon/on.png")

	go app.watchFiles()
	go app.watchIPN()
}

func (app *App) watchFiles() {

}

func (app *App) watchIPN() {
	for {
		watcher, err := app.client.WatchIPNBus(app.ctx, 0)
		if err != nil {
			log.Printf("loading IPN bus watcher: %s\n", err)
			time.Sleep(time.Second)
		}

		for {
			not, err := watcher.Next()
			if err != nil {
				log.Printf("Watching IPN Bus: %s\n", err)
				break
			}

			runtime.EventsEmit(app.ctx, "update_all")

			log.Printf("IPN bus update: %v\n", not)
		}
	}
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
	// TODO: get prefs, set prefs
	_ = peer
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

func (app *App) Self() Peer {
	status, err := app.client.Status(app.ctx)
	if err != nil {
		panic(err)
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

func (app *App) Peers() []Namespace {
	status, err := app.client.Status(app.ctx)
	if err != nil {
		panic(err)
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
