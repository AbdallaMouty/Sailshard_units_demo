package main

import (
    "log"
    "github.com/g3n/engine/app"
    "github.com/g3n/engine/camera"
    "github.com/g3n/engine/core"
    "github.com/g3n/engine/geometry"
    "github.com/g3n/engine/graphic"
    "github.com/g3n/engine/light"
    "github.com/g3n/engine/material"
    "github.com/g3n/engine/math32"
    "github.com/g3n/engine/renderer"
    "github.com/g3n/engine/gui"
    "github.com/g3n/engine/gls"
    "time"
    "fmt"
)

const WorldSize=16
const UnitSize=0.25

var units [WorldSize][WorldSize][WorldSize]bool
var voxelRoot *core.Node
var cubeGeom *geometry.Geometry
var matTop, matSide, matBottom *material.Standard
var fpsLabel *gui.Label
var frames int
var lastFPS time.Time

func main(){
    a:=app.App()
    scene:=core.NewNode()
    cam:=camera.New(1); cam.SetPosition(4,4,12)
    scene.Add(cam)
    scene.Add(light.NewAmbient(&math32.Color{1,1,1},0.6))
    dir:=light.NewDirectional(&math32.Color{1,1,0.95},1.0)
    dir.SetPosition(1,2,1); scene.Add(dir)
    fpsLabel=gui.NewLabel("FPS:0"); fpsLabel.SetPosition(6,6); scene.Add(fpsLabel)
    voxelRoot=core.NewNode(); scene.Add(voxelRoot)
    cubeGeom=geometry.NewBox(UnitSize,UnitSize,UnitSize)
    matTop=material.NewStandard(math32.NewColor("green"))
    matSide=material.NewStandard(math32.NewColor("saddlebrown"))
    matBottom=material.NewStandard(math32.NewColor("gray"))
    initWorld(); buildVoxels()

    last:=time.Now()
    a.Gls().ClearColor(0.25,0.3,0.36,1)
    a.Run(func(r *renderer.Renderer,dt time.Duration){
        frames++
        if time.Since(lastFPS)>time.Second{fpsLabel.SetText(fmt.Sprintf("FPS:%d",frames));frames=0;lastFPS=time.Now()}
        a.Gls().Clear(gls.DEPTH_BUFFER_BIT|gls.STENCIL_BUFFER_BIT|gls.COLOR_BUFFER_BIT)
        r.Render(scene,cam)
    })
}

func initWorld(){for x:=0;x<WorldSize;x++{for z:=0;z<WorldSize;z++{h:=8;for y:=0;y<WorldSize;y++{units[x][y][z]=y<=h}}}}
func isSolid(x,y,z int)bool{if x<0||y<0||z<0||x>=WorldSize||y>=WorldSize||z>=WorldSize{return false};return units[x][y][z]}
func exposed(x,y,z int)bool{if !units[x][y][z]{return false};dirs:=[6][3]int{{1,0,0},{-1,0,0},{0,1,0},{0,-1,0},{0,0,1},{0,0,-1}};for _,d:=range dirs{if !isSolid(x+d[0],y+d[1],z+d[2]){return true}};return false}
func buildVoxels(){voxelRoot.ClearChildren(true);for x:=0;x<WorldSize;x++{for y:=0;y<WorldSize;y++{for z:=0;z<WorldSize;z++{if !exposed(x,y,z){continue};mat:=matSide;if y==0{mat=matBottom};if y==WorldSize-1{mat=matTop};mesh:=graphic.NewMesh(cubeGeom,mat);mesh.SetPosition((float32(x)-WorldSize/2)*UnitSize,(float32(y)-WorldSize/2)*UnitSize,(float32(z)-WorldSize/2)*UnitSize);voxelRoot.Add(mesh)}}}}
