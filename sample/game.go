package main

import (
	"image"
	"log"
	"math/rand"

	_ "image/png"

	"golang.org/x/mobile/asset"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/exp/sprite"
	"golang.org/x/mobile/exp/sprite/clock"
)

const (
	tileWidth, tileHeight = 16, 16 // 各タイルの幅と高さ
	tilesX, tilesY        = 16, 16 // 水平方向のタイルの数

	gopherTile = 1 // gopher が描かれるタイル (0-indexed)

	groundChangeProb = 5                                  // 地面の高さが変わる確率（1/probabilityで変わる）
	groundMin        = tileHeight * (tilesY - 2*tilesY/5) // 地面の変化の最小値
	groundMax        = tileHeight * tilesY                // 地面の変化の最大値
	initGroundY      = tileHeight * (tilesY - 1)          // 地面のY座標
)

type Game struct {
	groundY  [tilesX + 3]float32 // 地面のX座標。tilesXの数+定数分のサイズのfloat32配列を用意する
	lastCalc clock.Time          // 最後にフレームを計算した時間
}

func NewGame() *Game {
	var g Game
	g.reset()
	return &g
}

// ゲームの初期化
func (g *Game) reset() {
	for i := range g.groundY {
		g.groundY[i] = initGroundY // 地面のX座標分初期化する。
	}
}

func (g *Game) Scene(eng sprite.Engine) *sprite.Node {
	texs := loadTextures(eng)

	scene = &sprite.Node{} // sceneのルートノードを生成
	eng.Register(scene)    // Engineオブジェクトにルートノードを登録
	// ルートの初期位置やスケールを設定する
	eng.SetTransform(scene, f32.Affine{
		{1, 0, 0},
		{0, 1, 0},
	})

	newNode := func(fn arrangerFunc) {
		n := &sprite.Node{Arranger: arrangerFunc(fn)}
		eng.Register(n)
		scene.AppendChild(n)
	}

	// The ground.
	// 地面を描画するメソッド
	for i := range g.groundY {
		i := i
		newNode(func(eng sprite.Engine, n *sprite.Node, t clock.Time) {
			eng.SetSubTex(n, texs[texGround]) //texGroundのテクスチャを使う
			eng.SetTransform(n, f32.Affine{
				{tileWidth, 0, float32(i) * tileWidth},
				{0, tileHeight, g.groundY[i]}, //地面を描画する
			})
		})
	}

	// The gopher.
	newNode(func(eng sprite.Engine, n *sprite.Node, t clock.Time) {
		eng.SetSubTex(n, texs[texGopher])
		eng.SetTransform(n, f32.Affine{
			{tileWidth, 0, tileWidth * gopherTile},
			{0, tileHeight, 0},
		})
	})

	return scene
}

type arrangerFunc func(e sprite.Engine, n *sprite.Node, t clock.Time)

func (a arrangerFunc) Arrange(e sprite.Engine, n *sprite.Node, t clock.Time) {
	a(e, n, t)
}

const (
	texGopher = iota
	texGround
)

func loadTextures(eng sprite.Engine) []sprite.SubTex {
	a, err := asset.Open("placeholder-sprites.png")
	if err != nil {
		log.Fatal(err)
	}
	defer a.Close() // 処理が終わったらCloseする

	m, _, err := image.Decode(a)
	if err != nil {
		log.Fatal(err)
	}

	t, err := eng.LoadTexture(m)
	if err != nil {
		log.Fatal(err)
	}

	const n = 128
	return []sprite.SubTex{
		texGopher: sprite.SubTex{t, image.Rect(1+0, 0, n-1, n)},     //splite画像の一番左の青色のテクスチャを切り出す
		texGround: sprite.SubTex{t, image.Rect(1+n*2, 0, n*3-1, n)}, //splite画像の左からn-1番目のテクスチャを切り出す
	}

}

func (g *Game) Update(now clock.Time) {
	// Compute game states up to now
	for ; g.lastCalc < now; g.lastCalc++ {
		g.calcFrame()
	}
}

func (g *Game) calcFrame() {
	// そのフレームでのゲーム状態を計算する
	g.calcScroll()
}

func (g *Game) calcScroll() {
	// 1秒あたり3つの地面タイルを作成する
	if g.lastCalc%20 == 0 {
		g.newGroundTile()
	}
}

func (g *Game) newGroundTile() {
	// 次の地面のオフセットを計算する
	next := g.nextGroundY()

	// 地面とを左に移動する
	copy(g.groundY[:], g.groundY[1:])
	g.groundY[len(g.groundY)-1] = next
}

func (g *Game) nextGroundY() float32 {

	prev := g.groundY[len(g.groundY)-1]
	if change := rand.Intn(groundChangeProb) == 0; change {
		return (groundMax-groundMin)*rand.Float32() + groundMin
	}
	return prev
}
