package tetris

import (
	"log"
	"time"

	"github.com/gdamore/tcell"
)

// EventEngineStopRun stop the run of the engine
type EventEngineStopRun struct {
	EventGame
}

// NewEngine creates new engine
func NewEngine() {
	engine = &Engine{
		chanStop:     make(chan struct{}),
		chanEventKey: make(chan *tcell.EventKey, 8),
		mode:         engineModeGameOver,
		tickTime:     time.Hour,
		ranking:      NewRanking(),
		ai:           NewAi(),
	}
	board.Clear()
	go engine.Run()
}

// Run runs the engine
func (engine *Engine) Run() {
	log.Println("Engine Run start")

	var event tcell.Event

loop:
	for {
		event = screen.PollEvent()
		switch eventType := event.(type) {
		case *tcell.EventKey:
			select {
			case engine.chanEventKey <- eventType:
			default:
			}
		case *EventEngineStopRun:
			break loop
		case *tcell.EventResize:
			view.RefreshScreen()
		default:
			logger.Printf("event type %T", eventType)
		}
	}

	log.Println("Engine Run end")
}

// Start the game
func (engine *Engine) Start() {
	log.Println("Engine Start start")

	engine.timer = time.NewTimer(engine.tickTime)
	engine.timer.Stop()
	engine.aiTimer = time.NewTimer(engine.tickTime)
	engine.aiTimer.Stop()

	view.RefreshScreen()

	var eventKey *tcell.EventKey

loop:
	for {
		select {
		case <-engine.chanStop:
			break loop
		default:
			select {
			case eventKey = <-engine.chanEventKey:
				engine.ProcessEventKey(eventKey)
				view.RefreshScreen()
			case <-engine.timer.C:
				engine.tick()
			case <-engine.aiTimer.C:
				engine.ai.ProcessQueue()
				engine.aiTimer.Reset(engine.tickTime / aiTickDivider)
			case <-engine.chanStop:
				break loop
			}
		}
	}

	screen.PostEventWait(&EventEngineStopRun{})

	log.Println("Engine Start end")
}

// Stop the game
func (engine *Engine) Stop() {
	log.Println("Engine Stop start")

	if !engine.stopped {
		engine.stopped = true
		engine.mode = engineModeStopped
		close(engine.chanStop)
	}
	engine.timer.Stop()
	engine.aiTimer.Stop()

	log.Println("Engine Stop end")
}

// Pause the game
func (engine *Engine) Pause() {
	if !engine.timer.Stop() {
		select {
		case <-engine.timer.C:
		default:
		}
	}
	if !engine.aiTimer.Stop() {
		select {
		case <-engine.aiTimer.C:
		default:
		}
	}
	engine.mode = engineModePaused
}

// UnPause the game
func (engine *Engine) UnPause() {
	engine.timer.Reset(engine.tickTime)
	if engine.aiEnabled {
		engine.aiTimer.Reset(engine.tickTime / aiTickDivider)
		engine.mode = engineModeRunWithAI
	} else {
		engine.mode = engineModeRun
	}
}

// PreviewBoard sets previewBoard to true
func (engine *Engine) PreviewBoard() {
	engine.mode = engineModePreview
}

// NewGame resets board and starts a new game
func (engine *Engine) NewGame() {
	log.Println("Engine NewGame start")

	board.Clear()
	engine.tickTime = 480 * time.Millisecond
	engine.score = 0
	engine.level = 1
	engine.deleteLines = 0

loop:
	for {
		select {
		case <-engine.chanEventKey:
		default:
			break loop
		}
	}

	if engine.aiEnabled {
		engine.ai.GetBestQueue()
		engine.mode = engineModeRunWithAI
	} else {
		engine.mode = engineModeRun
	}
	engine.UnPause()

	log.Println("Engine NewGame end")
}

// ResetTimer resets the time for lock delay or tick time
func (engine *Engine) ResetTimer(duration time.Duration) {
	if !engine.timer.Stop() {
		select {
		case <-engine.timer.C:
		default:
		}
	}
	if duration == 0 {
		// duration 0 means tick time
		engine.timer.Reset(engine.tickTime)
	} else {
		// duration is lock delay
		engine.timer.Reset(duration)
	}
}

// AiGetBestQueue calls AI to get best queue
func (engine *Engine) AiGetBestQueue() {
	if !engine.aiEnabled {
		return
	}
	go engine.ai.GetBestQueue()
}

// tick move mino down and refreshes screen
func (engine *Engine) tick() {
	board.MinoMoveDown()
	view.RefreshScreen()
}

// AddDeleteLines adds deleted lines to score
func (engine *Engine) AddDeleteLines(lines int) {
	engine.deleteLines += lines
	if engine.deleteLines > 999999 {
		engine.deleteLines = 999999
	}

	switch lines {
	case 1:
		engine.AddScore(40 * (engine.level + 1))
	case 2:
		engine.AddScore(100 * (engine.level + 1))
	case 3:
		engine.AddScore(300 * (engine.level + 1))
	case 4:
		engine.AddScore(1200 * (engine.level + 1))
	}

	if engine.level < engine.deleteLines/10 {
		engine.LevelUp()
	}
}

// AddScore adds to score
func (engine *Engine) AddScore(add int) {
	engine.score += add
	if engine.score > 9999999 {
		engine.score = 9999999
	}
}

// LevelUp goes up a level
func (engine *Engine) LevelUp() {
	if engine.level >= 30 {
		return
	}

	engine.level++
	switch {
	case engine.level > 29:
		engine.tickTime = 10 * time.Millisecond
	case engine.level > 25:
		engine.tickTime = 20 * time.Millisecond
	case engine.level > 19:
		// 50 to 30
		engine.tickTime = time.Duration(10*(15-engine.level/2)) * time.Millisecond
	case engine.level > 9:
		// 150 to 60
		engine.tickTime = time.Duration(10*(25-engine.level)) * time.Millisecond
	default:
		// 480 to 160
		engine.tickTime = time.Duration(10*(52-4*engine.level)) * time.Millisecond
	}
}

// GameOver pauses engine and sets to game over
func (engine *Engine) GameOver() {
	log.Println("Engine GameOver start")

	engine.Pause()
	engine.mode = engineModeGameOver

	view.ShowGameOverAnimation()

loop:
	for {
		select {
		case <-engine.chanEventKey:
		default:
			break loop
		}
	}

	engine.ranking.InsertScore(uint64(engine.score))
	engine.ranking.Save()

	log.Println("Engine GameOver end")
}

// EnabledAi enables the AI
func (engine *Engine) EnabledAi() {
	engine.aiEnabled = true
	go engine.ai.GetBestQueue()
	engine.aiTimer.Reset(engine.tickTime / aiTickDivider)
}

// DisableAi disables the AI
func (engine *Engine) DisableAi() {
	engine.aiEnabled = false
	engine.mode = engineModeRun
	if !engine.aiTimer.Stop() {
		select {
		case <-engine.aiTimer.C:
		default:
		}
	}
}

// EnabledEditMode enables edit mode
func (engine *Engine) EnabledEditMode() {
	edit.EnabledEditMode()
	engine.mode = engineModeEdit
}

// DisableEditMode disables edit mode
func (engine *Engine) DisableEditMode() {
	edit.DisableEditMode()
	engine.mode = engineModePreview
}
