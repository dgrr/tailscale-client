package main

import (
	"context"
	"log"

	"github.com/dgrr/tl"
	"github.com/gen2brain/beeep"
	"tailscale.com/client/tailscale"
	"tailscale.com/ipn"
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
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	beeep.Notify("Tailscale", "tailscale started", "icon/on.png")
}

func (a *App) CurrentAccount() string {
	current, _, err := a.client.ProfileStatus(a.ctx)
	if err != nil {
		panic(err)
	}

	return current.Name
}

func (a *App) Accounts() []string {
	current, all, err := a.client.ProfileStatus(a.ctx)
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

func (a *App) SwitchTo(account string) {
	current, all, err := a.client.ProfileStatus(a.ctx)
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

	a.client.SwitchProfile(a.ctx, all[0].ID)
}
