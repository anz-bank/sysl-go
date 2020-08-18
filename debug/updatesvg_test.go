package debug

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpdateSvg(t *testing.T) {
	out, err := UpdateSvg(testSvg, "{'Do'}", "green")

	fmt.Println(out)
	assert.NoError(t, err)
	assert.Contains(t, out, "green")
}

func TestUpdateSvg_Twice(t *testing.T) {
	out, err := UpdateSvg(testSvg, "{'Do'}", "green")
	out, err = UpdateSvg(out, "{'Do'}", "blue")

	assert.NoError(t, err)
	assert.Contains(t, out, "blue")
}

func TestUpdateSvg_Big(t *testing.T) {
	out, err := UpdateSvg(svg, "{'PaymentServer POST /pay <-- mastercard POST /pay'}", "green")

	assert.NoError(t, err)
	assert.Contains(t, out, "green")
}

const testSvg = `
<?xml version="1.0" encoding="UTF-8" standalone="no"?><svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" contentScriptType="application/ecmascript" contentStyleType="text/css" height="191px" preserveAspectRatio="none" style="width:76px;height:191px;" version="1.1" viewBox="0 0 76 191" width="76px" zoomAndPan="magnify"><defs><filter height="300%" id="f17zc7ifjf2he3" width="300%" x="-1" y="-1"><feGaussianBlur result="blurOut" stdDeviation="2.0"/><feColorMatrix in="blurOut" result="blurOut2" type="matrix" values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 .4 0"/><feOffset dx="4.0" dy="4.0" in="blurOut2" result="blurOut3"/><feBlend in="SourceGraphic" in2="blurOut3" mode="normal"/></filter></defs><g><rect fill="#FFFFFF" filter="url(#f17zc7ifjf2he3)" height="29.1328" style="stroke: #A80036; stroke-width: 1.0;" width="10" x="40" y="82.4297"/><line style="stroke: #A80036; stroke-width: 1.0; stroke-dasharray: 5.0,5.0;" x1="44.5" x2="44.5" y1="51.2969" y2="129.5625"/><text fill="#000000" font-family="sans-serif" font-size="14" lengthAdjust="spacingAndGlyphs" textLength="39" x="22.5" y="47.9951">Client</text><ellipse cx="45" cy="19" fill="#FEFECE" filter="url(#f17zc7ifjf2he3)" rx="12" ry="12" style="stroke: #A80036; stroke-width: 2.0;"/><polygon fill="#A80036" points="41,7,47,2,45,7,47,12,41,7" style="stroke: #A80036; stroke-width: 1.0;"/><text fill="#000000" font-family="sans-serif" font-size="14" lengthAdjust="spacingAndGlyphs" textLength="39" x="22.5" y="141.5576">Client</text><ellipse cx="45" cy="160.8594" fill="#FEFECE" filter="url(#f17zc7ifjf2he3)" rx="12" ry="12" style="stroke: #A80036; stroke-width: 2.0;"/><polygon fill="#A80036" points="41,148.8594,47,143.8594,45,148.8594,47,153.8594,41,148.8594" style="stroke: #A80036; stroke-width: 1.0;"/><rect fill="#FFFFFF" filter="url(#f17zc7ifjf2he3)" height="29.1328" style="stroke: #A80036; stroke-width: 1.0;" width="10" x="40" y="82.4297"/><polygon fill="#808080" points="28,78.4297,38,82.4297,28,86.4297,32,82.4297" style="stroke: #808080; stroke-width: 1.0;"/><line style="stroke: #808080; stroke-width: 1.0;" x1="3" x2="34" y1="82.4297" y2="82.4297"/><text fill="#000000" font-family="sans-serif" font-size="13" lengthAdjust="spacingAndGlyphs" textLength="18" x="10" y="77.3638">Do</text><polygon fill="#808080" points="14,107.5625,4,111.5625,14,115.5625,10,111.5625" style="stroke: #808080; stroke-width: 1.0;"/><line style="stroke: #808080; stroke-width: 1.0; stroke-dasharray: 2.0,2.0;" x1="8" x2="44" y1="111.5625" y2="111.5625"/><text fill="#000000" font-family="sans-serif" font-size="13" lengthAdjust="spacingAndGlyphs" textLength="15" x="20" y="106.4966">ok</text><!--MD5=[f34e79fc822bf6e90c0de8d2c1c3d0da]
@startuml
skinparam MaxMessageSize 250
skinparam sequence {
    ArrowColor grey
}

control "Client"

 -> Client : Do
activate Client
 <- - Client : ok
deactivate Client
@enduml

PlantUML version 1.2020.13(Sat Jun 13 12:26:38 UTC 2020)
(GPL source distribution)
Java Runtime: OpenJDK Runtime Environment
JVM: OpenJDK 64-Bit Server VM
Default Encoding: UTF-8
Language: en
Country: null
--></g></svg>
`
