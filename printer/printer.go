package printer

import (
	"fmt"
	"io"
	"math"
)

type Printer struct {
	LayerHeight      float64
	FlowCorrection   float64
	Temperature      float64
	MaxSpeed         float64
	CenterX          float64
	CenterY          float64
	FilamentDiameter float64
	TravelSpeed      float64
	PrintSpeed       float64
	RetractionSpeed  float64
	RetractionLength float64
	Output           io.Writer
	x, y, z, e       float64
}

func (p *Printer) SendCommand(format string, args ...interface{}) {
	io.WriteString(p.Output, fmt.Sprintf(format+"\n", args...))
}

func (p *Printer) Preamble() {
	p.SendCommand("G28       ; home all axis")
	p.SendCommand("G21       ; set units to millimeters")
	p.SendCommand("G90       ; set absolute coordinates")
	p.SendCommand("M82       ; use absolute distances for extrusion")
	p.SetTempAndWait(p.Temperature)
	p.SendCommand("")
}

func (p *Printer) Postamble() {
	p.retract()
	p.SendCommand("")
	p.SendCommand("M104 S0 ; turn off temperature")
	p.SendCommand("G28 X0  ; home X axis")
	p.SendCommand("M84     ; turn off motors")
}

func (p *Printer) Comment(format string, args ...interface{}) {
	commentFormat := fmt.Sprintf("\n; ------- %s ------", format)
	p.SendCommand(commentFormat, args...)
}

func (p *Printer) ZeroExtrusion() {
	p.SendCommand("G92 E0    ; zero extrusion")
	p.e = 0.0
}

func (p *Printer) SetTempAndWait(temp float64) {
	p.SendCommand("M109 S%.3f ; set and wait head temperature", temp)
}

func (p *Printer) Raise() {
	p.z += p.LayerHeight
	p.SendCommand("G0 Z%.3f F%.3f ; raise", p.z, 60*p.TravelSpeed)
	p.ZeroExtrusion()
}

func (p *Printer) Move(x, y float64) {
	p.x = x + p.CenterX
	p.y = y + p.CenterY
	p.SendCommand("G0 X%.3f Y%.3f F%.3f ; move", p.x, p.y, 60*p.TravelSpeed)
}

func (p *Printer) MoveAndRetract(x, y float64) {
	p.retract()
	p.Move(x, y)
	p.unretract()
}

func (p *Printer) Print(x, y, spread float64) {
	tx := x + p.CenterX
	ty := y + p.CenterY
	dx := tx - p.x
	dy := ty - p.y
	p.x = tx
	p.y = ty
	p.e += p.getExtrusionLength(p.linearDistance(dx, dy), spread)
	p.SendCommand("G0 X%.3f Y%.3f E%.3f F%.3f ; print", p.x, p.y, p.e, 60*p.PrintSpeed)
}

func (p *Printer) getExtrusionLength(d, spread float64) float64 {
	return 4 * p.FlowCorrection * p.LayerHeight * p.LayerHeight * d * (spread + math.Pi/2) / (math.Pi * p.FilamentDiameter * p.FilamentDiameter)
}

func (p *Printer) linearDistance(dx, dy float64) float64 {
	return math.Sqrt(dx*dx + dy*dy)
}

func (p *Printer) retract() {
	p.e -= p.RetractionLength
	p.SendCommand("G0 E%.3f F%.3f ; retract", p.e, 60*p.RetractionSpeed)
}

func (p *Printer) unretract() {
	p.e += p.RetractionLength
	p.SendCommand("G0 E%.3f F%.3f ; unretract", p.e, 60*p.RetractionSpeed)
}
