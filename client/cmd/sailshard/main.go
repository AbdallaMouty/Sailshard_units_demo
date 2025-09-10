package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/window"
)

const WorldSize = 16
const UnitSize = 1.0

var (
	units                      [WorldSize][WorldSize][WorldSize]bool
	voxelRoot                  *core.Node
	cubeGeom                   *geometry.Geometry
	matTop, matSide, matBottom *material.Standard
	fpsLabel                   *gui.Label
	frames                     int
	lastFPS                    time.Time

	// Controls
	yaw, pitch                float32
	moveForward, moveBackward bool
	moveLeft, moveRight       bool

	// Mouse tracking
	lastMouseX, lastMouseY float32
	firstMouse             = true
)

// TCP messages
type TickMsg struct {
	Type string `json:"type"`
	Tick uint64 `json:"tick"`
	Ts   int64  `json:"ts"`
}

type EchoMsg struct {
	Type string `json:"type"`
	From string `json:"from"`
	Body string `json:"body"`
	Ts   int64  `json:"ts"`
}

func main() {
	a := app.App()
	scene := core.NewNode()

	// ---------------- Camera ----------------
	cam := camera.New(1)
	cam.SetPosition(0, 0, 3)
	scene.Add(cam)

	// Set up orbit control for the camera
	camera.NewOrbitControl(cam)

	onResize := func(evname string, ev interface{}) {
		// Get framebuffer size and update viewport accordingly
		width, height := a.GetSize()
		a.Gls().Viewport(0, 0, int32(width), int32(height))
		// Update the camera's aspect ratio
		cam.SetAspect(float32(width) / float32(height))
	}
	a.Subscribe(window.OnWindowSize, onResize)
	onResize("", nil)

	// ---------------- Lighting ----------------
	scene.Add(light.NewAmbient(math32.NewColor("white"), 0.6))
	dir := light.NewDirectional(math32.NewColor("white"), 1.0)
	dir.SetPosition(1, 2, 1)
	scene.Add(dir)

	// ---------------- GUI ----------------
	fpsLabel = gui.NewLabel("FPS:0")
	fpsLabel.SetPosition(6, 6)
	scene.Add(fpsLabel)

	// ---------------- Voxels ----------------
	voxelRoot = core.NewNode()
	scene.Add(voxelRoot)

	cubeGeom = geometry.NewBox(UnitSize, UnitSize, UnitSize)
	matTop = material.NewStandard(math32.NewColor("green"))
	matSide = material.NewStandard(math32.NewColor("saddlebrown"))
	matBottom = material.NewStandard(math32.NewColor("gray"))

	initWorld()
	buildVoxels()

	// ---------------- Controls ----------------
	initControls(a, cam)

	// ---------------- TCP Connection ----------------
	conn, err := net.Dial("tcp", "127.0.0.1:27015")
	if err != nil {
		log.Println("Error connecting to server:", err)
	} else {
		go handleServer(conn)
	}

	// ---------------- Render loop ----------------
	a.Gls().ClearColor(0.25, 0.3, 0.36, 1)
	a.Run(func(r *renderer.Renderer, dt time.Duration) {
		frames++
		if time.Since(lastFPS) > time.Second {
			fpsLabel.SetText(fmt.Sprintf("FPS:%d", frames))
			fmt.Printf("FPS: %d | Camera: %v\n", frames, cam.Position())
			frames = 0
			lastFPS = time.Now()
		}

		updateCamera(cam, float32(dt.Seconds()))

		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		r.Render(scene, cam)
	})
}

// ---------------- World ----------------

func initWorld() {
	for x := 0; x < WorldSize; x++ {
		for z := 0; z < WorldSize; z++ {
			h := 8
			for y := 0; y < WorldSize; y++ {
				units[x][y][z] = y <= h
			}
		}
	}
}

func isSolid(x, y, z int) bool {
	if x < 0 || y < 0 || z < 0 || x >= WorldSize || y >= WorldSize || z >= WorldSize {
		return false
	}
	return units[x][y][z]
}

func exposed(x, y, z int) bool {
	if !units[x][y][z] {
		return false
	}
	dirs := [6][3]int{{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1}}
	for _, d := range dirs {
		if !isSolid(x+d[0], y+d[1], z+d[2]) {
			return true
		}
	}
	return false
}

func clearChildren(n *core.Node) {
	for _, child := range n.Children() {
		n.Remove(child)
	}
}

func buildVoxels() {
	clearChildren(voxelRoot)
	offset := float32(WorldSize) / 2
	for x := 0; x < WorldSize; x++ {
		for y := 0; y < WorldSize; y++ {
			for z := 0; z < WorldSize; z++ {
				if !exposed(x, y, z) {
					continue
				}
				mat := matSide
				if y == 0 {
					mat = matBottom
				}
				if y == WorldSize-1 {
					mat = matTop
				}
				mesh := graphic.NewMesh(cubeGeom, mat)
				mesh.SetPosition(
					float32(x)-offset,
					float32(y),
					float32(z)-offset,
				)
				voxelRoot.Add(mesh)
			}
		}
	}
}

// ---------------- Controls ----------------

func initControls(a *app.Application, cam *camera.Camera) {
	win := a

	// Mouse look
	win.Subscribe(window.OnCursor, func(evname string, ev interface{}) {
		mev := ev.(*window.CursorEvent)
		x := float32(mev.Xpos)
		y := float32(mev.Ypos)

		if firstMouse {
			lastMouseX = x
			lastMouseY = y
			firstMouse = false
		}

		xoffset := x - lastMouseX
		yoffset := lastMouseY - y
		lastMouseX = x
		lastMouseY = y

		sensitivity := float32(0.002)
		yaw += xoffset * sensitivity
		pitch += yoffset * sensitivity

		if pitch > 1.5 {
			pitch = 1.5
		}
		if pitch < -1.5 {
			pitch = -1.5
		}
	})

	// Keyboard input
	win.Subscribe(window.OnKeyDown, func(evname string, ev interface{}) {
		kev := ev.(*window.KeyEvent)
		switch kev.Key {
		case window.KeyW:
			moveForward = true
		case window.KeyS:
			moveBackward = true
		case window.KeyA:
			moveLeft = true
		case window.KeyD:
			moveRight = true
		}
	})
	win.Subscribe(window.OnKeyUp, func(evname string, ev interface{}) {
		kev := ev.(*window.KeyEvent)
		switch kev.Key {
		case window.KeyW:
			moveForward = false
		case window.KeyS:
			moveBackward = false
		case window.KeyA:
			moveLeft = false
		case window.KeyD:
			moveRight = false
		}
	})
}

func updateCamera(cam *camera.Camera, dt float32) {
	speed := float32(10.0)
	velocity := speed * dt

	dirVec := math32.NewVector3(
		math32.Cos(pitch)*math32.Sin(yaw),
		math32.Sin(pitch),
		math32.Cos(pitch)*math32.Cos(yaw),
	)

	pos := cam.Position()

	if moveForward {
		pos.Add(dirVec.Clone().MultiplyScalar(velocity))
	}
	if moveBackward {
		pos.Sub(dirVec.Clone().MultiplyScalar(velocity))
	}
	if moveLeft {
		right := math32.NewVector3(dirVec.Z, 0, -dirVec.X)
		right.Normalize()
		pos.Add(right.MultiplyScalar(velocity))
	}
	if moveRight {
		left := math32.NewVector3(-dirVec.Z, 0, dirVec.X)
		left.Normalize()
		pos.Add(left.MultiplyScalar(velocity))
	}

	cam.SetPositionVec(&pos)
	lookAt := pos.Clone().Add(dirVec)
	cam.LookAt(lookAt, math32.NewVector3(0, 1, 0))
}

// ---------------- TCP ----------------

func handleServer(conn net.Conn) {
	defer conn.Close()
	sc := bufio.NewScanner(conn)
	for sc.Scan() {
		line := sc.Text()
		var tick TickMsg
		var echo EchoMsg

		if err := json.Unmarshal([]byte(line), &tick); err == nil && tick.Type == "tick" {
			fmt.Printf("Tick %d @ %d\n", tick.Tick, tick.Ts)
			continue
		}

		if err := json.Unmarshal([]byte(line), &echo); err == nil && echo.Type == "echo" {
			fmt.Printf("[%s]: %s\n", echo.From, echo.Body)
			continue
		}
	}
}
