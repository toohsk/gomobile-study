package main

import (
	"time"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/exp/sprite"
	"golang.org/x/mobile/exp/sprite/clock"
	"golang.org/x/mobile/exp/sprite/glsprite"
	"golang.org/x/mobile/gl"
)

func main() {
	app.Main(func(a app.App) {
		var glctx gl.Context
		/**
		 sizeについて
		 - 画面に変更があった場合のEvent
		 - アプリ起動時に1度はpaint.Eventが発生する前に必ず呼ばれる
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
				if glctx == nil || e.External {
					continue
				}
				onPaint(glctx, sz)
				a.Publish()
				a.Send(paint.Event{}) // keep animating
			}
		}
	})
}

var (
	startTime = time.Now()
	images    *glutil.Images
	eng       sprite.Engine
	scene     *sprite.Node
)

// 開始時点で呼ばれる関数なので、Sprite Engineを初期化する
func onStart(glctx gl.Context) {
	images = glutil.NewImages(glctx) //OpenGLのContextからImageオブジェクトを生成
	eng = glsprite.Engine(images)    // Engineオブジェクトを生成する
	scene = &sprite.Node{}           // sceneのルートノードを生成
	eng.Register(scene)              // Engineオブジェクトにルートノードを登録
	// ルートの初期位置やスケールを設定する
	eng.SetTransform(scene, f32.Affine{
		{1, 0, 0},
		{0, 1, 0},
	})
}

// エンジンオブジェクトとテクスチャを破棄する
func onStop() {
	eng.Release()
	images.Release()
}

func onPaint(glctx gl.Context, sz size.Event) {
	glctx.ClearColor(1, 1, 1, 1)                                // クリアする色
	glctx.Clear(gl.COLOR_BUFFER_BIT)                            // 塗りつぶし？
	now := clock.Time(time.Since(startTime) * 60 / time.Second) // 60FramePerSecで計算し、現在のフレームを計算する
	eng.Render(scene, now, sz)
}
