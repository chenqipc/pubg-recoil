package game

import (
	"image"
	"time"

	"github.com/huuhait/pubg-recoil/pkg/gui"
	"github.com/huuhait/pubg-recoil/pkg/stats"
	"github.com/huuhait/pubg-recoil/pkg/utils"
	"github.com/zsmartex/pkg/v2/log"

	hook "github.com/robotn/gohook"
)

type Game struct {
	GUI         *gui.GUI
	recoil      *Recoil
	playerStats *stats.PlayerStats
	scanning    bool
}

func NewGame(GUI *gui.GUI) *Game {
	playerStats := stats.NewPlayerStats()

	return &Game{
		playerStats: playerStats,
		recoil:      NewRecoil(playerStats),
		GUI:         GUI,
	}
}

func (g *Game) ScanStandState() {
	// 默认设置为站立
	g.playerStats.SetStandState(stats.StandStateStand)
	g.GUI.SetStandState(stats.StandStateStand)
	/*for {
		if !g.scanning && g.playerStats.ReadyRecoil() {
			state, err := g.getStandState()
			if err != nil {
				log.Errorf("failed to scan stand state: %v", err)
				continue
			}
			g.playerStats.SetStandState(state)
			g.GUI.SetStandState(state)
		}
		time.Sleep(50 * time.Millisecond)
	}*/
}

func (g *Game) ScanBullets() {
	for {
		_, found := g.playerStats.GetActiveWeapon()

		if !g.scanning && g.playerStats.Hold().Fire && g.playerStats.Hold().Aim && found {
			availableBullets, err := g.isBulletsAvailable()
			if err != nil {
				log.Errorf("failed to scan bullets: %v", err)
				continue
			}

			g.playerStats.SetAvailableBullets(availableBullets)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func (g *Game) ScanInventory() {
	for {
		if !g.scanning {
			time.Sleep(20 * time.Millisecond)
			continue
		}

		screenshot, err := utils.Screenshot(image.Rect(0, 0, 2560, 1440))
		if err != nil {
			log.Errorf("failed to screenshot while scanning: %v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		isInventoryOpening, err := g.isInventoryOpening(screenshot)
		if err != nil {
			log.Errorf("failed to scan inventory: %v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if !isInventoryOpening {
			g.scanning = false
			continue
		}

		weapons, err := g.getWeapons(screenshot)
		if err != nil {
			log.Errorf("failed to scan weapons: %v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		g.playerStats.SetWeapons(weapons)
		g.GUI.SetWeapons(weapons)
		currentActiveWeapon, found := g.playerStats.GetActiveWeapon()
		if found {
			g.GUI.SetActiveWeapon(currentActiveWeapon)
		}

		time.Sleep(20 * time.Millisecond)
	}
}

// 全局变量用于存储Ctrl按键的状态
var ctrlPressed = false

func (g *Game) DeviceHook() {
	hook.Register(hook.KeyDown, []string{"tab"}, func(e hook.Event) {
		g.scanning = true
	})

	// 监听Ctrl按键按下和松开
	hook.Register(hook.KeyDown, []string{"ctrl"}, func(e hook.Event) {
		ctrlPressed = true
	})
	hook.Register(hook.KeyUp, []string{"ctrl"}, func(e hook.Event) {
		ctrlPressed = false
	})

	hook.Register(hook.KeyDown, []string{"1"}, func(e hook.Event) {
		// 如果Ctrl键已按下，则跳过该事件
		if ctrlPressed {
			return
		}
		g.playerStats.SetActiveWeapon(1)
		currentActiveWeapon, _ := g.playerStats.GetActiveWeapon()
		g.GUI.SetActiveWeapon(currentActiveWeapon)
	})

	hook.Register(hook.KeyDown, []string{"2"}, func(e hook.Event) {
		// 如果Ctrl键已按下，则跳过该事件
		if ctrlPressed {
			return
		}
		g.playerStats.SetActiveWeapon(2)
		currentActiveWeapon, _ := g.playerStats.GetActiveWeapon()
		g.GUI.SetActiveWeapon(currentActiveWeapon)
	})

	hook.Register(hook.KeyDown, []string{"c"}, func(e hook.Event) {
		cStats := g.playerStats.GetStandState()
		newStats := stats.StandStateSit
		if cStats == stats.StandStateStand || cStats == stats.StandStateLie {
			newStats = stats.StandStateSit
		} else if cStats == stats.StandStateSit {
			newStats = stats.StandStateStand
		}
		g.playerStats.SetStandState(newStats)
		g.GUI.SetStandState(newStats)
	})

	hook.Register(hook.KeyDown, []string{"z"}, func(e hook.Event) {
		cStats := g.playerStats.GetStandState()
		var newStats stats.StandState
		if cStats == stats.StandStateStand || cStats == stats.StandStateSit {
			newStats = stats.StandStateLie
		} else if cStats == stats.StandStateLie {
			newStats = stats.StandStateStand
		}
		g.playerStats.SetStandState(newStats)
		g.GUI.SetStandState(newStats)
	})

	hook.Register(hook.KeyDown, []string{"space"}, func(e hook.Event) {
		log.Info("空格")
		g.playerStats.SetStandState(stats.StandStateStand)
		g.GUI.SetStandState(stats.StandStateStand)
	})

	hook.Register(hook.MouseHold, []string{}, func(e hook.Event) {
		if e.Button == hook.MouseMap["right"] {
			g.playerStats.SetAim(true)
		} else if e.Button == hook.MouseMap["left"] {
			g.playerStats.SetFire(true)
		}
	})

	hook.Register(hook.MouseUp, []string{}, func(e hook.Event) {
		if e.Button == hook.MouseMap["right"] {
			g.playerStats.SetAim(false)
		} else if e.Button == hook.MouseMap["left"] {
			g.playerStats.SetFire(false)
		}
	})

	hook.Register(hook.MouseDown, []string{}, func(e hook.Event) {
		if e.Button == hook.MouseMap["right"] {
			g.playerStats.SetAim(false)
		} else if e.Button == hook.MouseMap["left"] {
			g.playerStats.SetFire(false)
		}
	})

	s := hook.Start()
	<-hook.Process(s)
}

func (g *Game) Start() {
	go g.ScanInventory()
	go g.ScanBullets()
	g.ScanStandState()
	go g.recoil.Start()
	go g.DeviceHook()
}
