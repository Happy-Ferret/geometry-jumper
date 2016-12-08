package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime/pprof"

	"geometry-jumper/game"
	"geometry-jumper/keyboard"
	"geometry-jumper/menu"
	"geometry-jumper/ranchblt"

	"time"

	"strings"

	"github.com/hajimehoshi/ebiten"
)

var (
	player          *game.PlayerCharacter
	keyboardWrapper = keyboard.NewKeyboardWrapper()
	shapeCollection *game.ShapeCollection
	logoScreen      *ranchblt.Logo
	showLogo        = true
	showMenu        = true
	mainMenu        menu.Menu
)

// Version is autoset from the build script
var Version string

// Build is autoset from the build script
var Build string

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func gameLoop(screen *ebiten.Image) error {
	if ebiten.IsRunningSlowly() {
		if game.Debug {
			go fmt.Println("slow")
		}
		return nil
	}

	keyboardWrapper.Update()

	go logoTimer()

	if showLogo && !game.Debug {
		logoScreen.Draw(screen)
		return nil
	}

	if showMenu {
		mainMenu.Update()
		mainMenu.Draw(screen)
		if keyboardWrapper.IsKeyPressed(ebiten.KeyEnter) {
			if strings.ToLower(mainMenu.Selected()) == "start" {
				showMenu = false
			} else if strings.ToLower(mainMenu.Selected()) == "exit" {
				return errors.New("User wanted to quit")
			}
		}
		return nil
	}

	if game.Debug {
		screen.DrawImage(game.UpperTrackLine, game.UpperTrackOpts)
		screen.DrawImage(game.LowerTrackLine, game.LowerTrackOpts)
	}

	if !player.Collided {
		shapeCollection.Update()
		player.Update()
	} else {
		shapeCollection.Stop = true
	}

	shapeCollection.Draw(screen)
	player.Draw(screen)

	go player.CheckCollision(shapeCollection)

	//ebitenutil.DebugPrint(screen, "Hello world!")

	if keyboardWrapper.KeyPushed(ebiten.KeyEscape) {
		return errors.New("User wanted to quit") //Best way to do this?
	}

	return nil
}

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)

		if err != nil {
			panic(err)
		}

		pprof.StartCPUProfile(f)

		defer pprof.StopCPUProfile()
	}

	game.Load()

	options := []*menu.Option{}
	options = append(options, &menu.Option{
		Text: "Start",
	})
	options = append(options, &menu.Option{
		Text: "Exit",
	})

	mainMenu = &menu.Regular{
		BackgroundImage: game.TitleImage,
		Height:          game.ScreenHeight,
		Width:           game.ScreenWidth,
		KeyboardWrapper: keyboardWrapper,
		Options:         options,
		Font:            game.Font,
	}

	square := game.NewSpawnDefaultSpeed(game.SquareType, game.LowerTrack)
	squareTwo := game.NewSpawnDefaultSpeed(game.SquareType, game.LowerTrack)
	triangle := game.NewSpawnDefaultSpeed(game.TriangleType, game.UpperTrack)
	circle := game.NewSpawnDefaultSpeed(game.CircleType, game.UpperTrack)

	firstGroup := game.NewSpawnGroup([]*game.Spawn{square}, 2500)
	secondGroup := game.NewSpawnGroup([]*game.Spawn{triangle, squareTwo}, 5000)
	thirdGroup := game.NewSpawnGroup([]*game.Spawn{circle}, 7500)
	pattern := game.NewPattern([]*game.SpawnGroup{firstGroup, secondGroup, thirdGroup})
	patternCollection := &game.PatternCollection{
		Patterns: map[int][]*game.Pattern{
			game.LowDifficulty: []*game.Pattern{pattern},
		},
	}

	shapeCollection = game.NewShapeCollection(patternCollection)

	player = game.NewPlayerCharacter("Test", game.PersonStandingImage, game.PersonJumpingImage, keyboardWrapper)

	logoScreen = ranchblt.NewLogoScreen(game.ScreenWidth, game.ScreenHeight)

	go fmt.Printf("Starting up game. Version %s, Build %s", Version, Build)

	ebiten.Run(gameLoop, game.ScreenWidth, game.ScreenHeight, 2, "Hello world!")
}

func logoTimer() {
	timer := time.NewTimer(time.Second * 2)
	<-timer.C
	showLogo = false
}
