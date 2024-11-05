package game

import (
	"fmt"
	"image"
	"image/png"
	"os"
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

		// 创建文件
		file, err := os.Create("screenshot.png")
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}
		defer file.Close()

		// 将图像编码为PNG并写入文件
		err = png.Encode(file, screenshot)
		if err != nil {
			fmt.Println("Error encoding PNG:", err)
			return
		}

		isInventoryOpening, err := g.isInventoryOpening(screenshot)
		if err != nil {
			log.Errorf("failed to scan inventory: %v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if !isInventoryOpening {
			log.Info("没有匹配到")
			g.scanning = false
			continue
		}
		log.Info("匹配到", isInventoryOpening)

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

func (g *Game) DeviceHook() {
	hook.Register(hook.KeyDown, []string{"tab"}, func(e hook.Event) {
		log.Info("开始扫描")
		g.scanning = true
	})

	hook.Register(hook.KeyDown, []string{"1"}, func(e hook.Event) {
		g.playerStats.SetActiveWeapon(1)
		currentActiveWeapon, _ := g.playerStats.GetActiveWeapon()
		g.GUI.SetActiveWeapon(currentActiveWeapon)
	})

	hook.Register(hook.KeyDown, []string{"2"}, func(e hook.Event) {
		g.playerStats.SetActiveWeapon(2)
		currentActiveWeapon, _ := g.playerStats.GetActiveWeapon()
		g.GUI.SetActiveWeapon(currentActiveWeapon)
	})

	hook.Register(hook.KeyDown, []string{"c"}, func(e hook.Event) {
		log.Info("c")
		cStats := g.playerStats.GetStandState()
		log.Info("原状态", cStats)
		newStats := stats.StandStateSit
		if cStats == stats.StandStateStand || cStats == stats.StandStateLie {
			newStats = stats.StandStateSit
			log.Info("新状态", newStats)
		} else if cStats == stats.StandStateSit {
			newStats = stats.StandStateStand
		}
		g.playerStats.SetStandState(newStats)
		g.GUI.SetStandState(newStats)
	})

	hook.Register(hook.KeyDown, []string{"z"}, func(e hook.Event) {
		log.Info("z")
		cStats := g.playerStats.GetStandState()
		var newStats stats.StandState
		if cStats == stats.StandStateStand || cStats == stats.StandStateSit {
			newStats = stats.StandStateLie
		} else if cStats == stats.StandStateLie {
			newStats = stats.StandStateStand
		}
		g.playerStats.SetStandState(newStats)
		//g.GUI.SetStandState(newStats)
	})

	hook.Register(hook.KeyDown, []string{"space"}, func(e hook.Event) {
		log.Info("空格")
		g.playerStats.SetStandState(stats.StandStateStand)
		//g.GUI.SetStandState(stats.StandStateStand)
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
