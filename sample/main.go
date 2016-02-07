package main

import (
	"math/rand"
	"time"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/exp/sprite"
	"golang.org/x/mobile/exp/sprite/clock"
	"golang.org/x/mobile/exp/sprite/glsprite"
	"golang.org/x/mobile/gl"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	app.Main(func(a app.App) {
		var glctx gl.Context
		/**
		 sizeについて
		 - 画面に変更があった場合のEvent
		 - アプリ起動時にpaint.Eventが発生する前に1度必ず呼ばれる
		 - サイズ情報が格納されている
		**/
		var sz size.Event
		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					// App visible
					glctx, _ = e.DrawContext.(gl.Context)
					onStart(glctx)
					a.Send(paint.Event{})
				case lifecycle.CrossOff:
					// App no longer visible.
					onStop()
					glctx = nil
				}
			case size.Event:
				sz = e
			case paint.Event:
				// OpenGLとOSからの描画イベントは無視する
				if glctx == nil || e.External {
					continue
				}
				// シーングラフを構築し、描画
				onPaint(glctx, sz)
				// 最終的に画面に出力
				a.Publish()
				a.Send(paint.Event{}) // keep animating
			case touch.Event: // タッチイベント
				if down := e.Type == touch.TypeBegin; down || e.Type == touch.TypeEnd {
					game.Press(down)
				}
			case key.Event: // キーボードイベント
				if e.Code != key.CodeSpacebar { // スペースキーのみをイベントとする
					break
				}
				if down := e.Direction == key.DirPress; down || e.Direction == key.DirRelease {
					game.Press(down)
				}
			}
		}
	})
}

var (
	startTime = time.Now()
	images    *glutil.Images
	eng       sprite.Engine
	scene     *sprite.Node
	game      *Game
)

// 開始時点で呼ばれる関数なので、Sprite Engineを初期化する
func onStart(glctx gl.Context) {
	images = glutil.NewImages(glctx) //OpenGLのContextからImageオブジェクトを生成
	eng = glsprite.Engine(images)    // Engineオブジェクトを生成する
	game = NewGame()
	scene = game.Scene(eng)
}

// エンジンオブジェクトとテクスチャを破棄する
func onStop() {
	eng.Release()
	images.Release()
	game = nil
}

func onPaint(glctx gl.Context, sz size.Event) {
	glctx.ClearColor(1, 1, 1, 1)                                // クリアする色
	glctx.Clear(gl.COLOR_BUFFER_BIT)                            // 塗りつぶし？
	now := clock.Time(time.Since(startTime) * 60 / time.Second) // 60FramePerSecで計算し、現在のフレームを計算する
	game.Update(now)
	eng.Render(scene, now, sz)
}
