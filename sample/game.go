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

	initScrollV = 1     // 垂直方向の初期値
	scrollA     = 0.001 // 加速度
	gravity     = 0.1   // 重力

	groundChangeProb = 5                                  // 地面の高さが変わる確率（1/probabilityで変わる）
	groundMin        = tileHeight * (tilesY - 2*tilesY/5) // 地面の変化の最小値
	groundMax        = tileHeight * tilesY                // 地面の変化の最大値
	initGroundY      = tileHeight * (tilesY - 1)          // 地面のY座標
)

type Game struct {
	gopher struct {
		y float32 // 垂直方向のオフセット
		v float32 // 速度
	}
	scroll struct {
		x float32 // 水平方向のオフセット
		v float32 // 速度
	}
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
	g.gopher.y = 0
	g.gopher.v = 0
	g.scroll.x = 0
	g.scroll.v = initScrollV
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
		// 地表の描画
		newNode(func(eng sprite.Engine, n *sprite.Node, t clock.Time) {
			eng.SetSubTex(n, texs[texGround]) //texGroundのテクスチャを使う
			eng.SetTransform(n, f32.Affine{
				{tileWidth, 0, float32(i)*tileWidth - g.scroll.x},
				{0, tileHeight, g.groundY[i]}, //地面を描画する
			})
		})
		// 地中の描画
		newNode(func(eng sprite.Engine, n *sprite.Node, t clock.Time) {
			eng.SetSubTex(n, texs[texEarth])
			eng.SetTransform(n, f32.Affine{
				{tileWidth, 0, float32(i) * tileWidth},
				{0, tileHeight * tilesY, g.groundY[i] + tileHeight},
			})
		})
	}

	// The gopher.
	newNode(func(eng sprite.Engine, n *sprite.Node, t clock.Time) {
		eng.SetSubTex(n, texs[texGopher])
		eng.SetTransform(n, f32.Affine{
			{tileWidth, 0, tileWidth * gopherTile},
			{0, tileHeight, g.gopher.y},
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
	texEarth
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
		texEarth:  sprite.SubTex{t, image.Rect(1+n*5, 0, n*6-1, n)}, //splite画像の左からn-1番目のテクスチャを切り出す
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
	g.calcGopher()
	g.calcScroll()
}

func (g *Game) calcScroll() {
	// 垂直方向の計算
	g.scroll.v += scrollA

	// オフセットの計算
	g.scroll.x += g.scroll.v

	// 新しい地面が必要な場合作成する
	for g.scroll.x > tileWidth {
		g.newGroundTile()
	}
}

func (g *Game) calcGopher() {
	// 速度を計算
	g.gopher.v += gravity

	// オフセットを計算
	g.gopher.y += g.gopher.v

	g.clampToGround()
}

func (g *Game) newGroundTile() {
	// 次の地面のオフセットを計算する
	next := g.nextGroundY()

	// 地面とを左に移動する
	g.scroll.x -= tileWidth
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

func (g *Game) clampToGround() {
	// gopherが立っている地面の位置を計算する
	minY := g.groundY[gopherTile]
	if y := g.groundY[gopherTile+1]; y < minY {
		minY = y
	}

	// gopherが地中に抜けるのを防ぐ
	maxGopherY := minY - tileHeight
	if g.gopher.y >= maxGopherY {
		g.gopher.v = 0
		g.gopher.y = maxGopherY
	}

}
