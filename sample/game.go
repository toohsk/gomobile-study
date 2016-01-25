package main

import (
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/exp/sprite"
)

type Game struct {
}

func NewGame() *Game {
	var g Game
	return &g
}

func (g *Game) Scene(eng sprite.Engine) *sprite.Node {
	scene = &sprite.Node{} // sceneのルートノードを生成
	eng.Register(scene)    // Engineオブジェクトにルートノードを登録
	// ルートの初期位置やスケールを設定する
	eng.SetTransform(scene, f32.Affine{
		{1, 0, 0},
		{0, 1, 0},
	})
	return scene
}
