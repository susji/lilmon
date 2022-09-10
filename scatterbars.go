package main

//
// This code is mostly based on `scatter.go` from gonum/plot.
//
// License from gonum:
//
// Copyright ©2013 The Gonum Authors. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//     * Redistributions of source code must retain the above copyright
//       notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above copyright
//       notice, this list of conditions and the following disclaimer in the
//       documentation and/or other materials provided with the distribution.
//     * Neither the name of the Gonum project nor the names of its authors and
//       contributors may be used to endorse or promote products derived from this
//       software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Copyright ©2015 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

type ScatterBars struct {
	plotter.XYs
	GlyphStyleFunc func(int) draw.GlyphStyle
	draw.GlyphStyle
	draw.LineStyle
}

func NewScatterBars(xys plotter.XYer) (*ScatterBars, error) {
	data, err := plotter.CopyXYs(xys)
	if err != nil {
		return nil, err
	}
	return &ScatterBars{
		XYs:        data,
		GlyphStyle: plotter.DefaultGlyphStyle,
		LineStyle:  plotter.DefaultLineStyle,
	}, err
}

func (pts *ScatterBars) Plot(c draw.Canvas, plt *plot.Plot) {
	trX, trY := plt.Transforms(&c)
	glyph := func(i int) draw.GlyphStyle { return pts.GlyphStyle }
	if pts.GlyphStyleFunc != nil {
		glyph = pts.GlyphStyleFunc
	}
	for i, p := range pts.XYs {
		pp := vg.Point{X: trX(p.X), Y: trY(p.Y)}
		p0 := (vg.Point{X: trX(p.X), Y: trY(0)})
		clipped := c.ClipLinesXY([]vg.Point{pp, p0})
		c.StrokeLines(pts.LineStyle, clipped...)
		c.DrawGlyph(glyph(i), pp)
	}
}

func (pts *ScatterBars) DataRange() (xmin, xmax, ymin, ymax float64) {
	return plotter.XYRange(pts)
}

func (pts *ScatterBars) GlyphBoxes(plt *plot.Plot) []plot.GlyphBox {
	glyph := func(i int) draw.GlyphStyle { return pts.GlyphStyle }
	if pts.GlyphStyleFunc != nil {
		glyph = pts.GlyphStyleFunc
	}
	bs := make([]plot.GlyphBox, len(pts.XYs))
	for i, p := range pts.XYs {
		bs[i].X = plt.X.Norm(p.X)
		bs[i].Y = plt.Y.Norm(p.Y)
		r := glyph(i).Radius
		bs[i].Rectangle = vg.Rectangle{
			Min: vg.Point{X: -r, Y: -r},
			Max: vg.Point{X: +r, Y: +r},
		}
	}
	return bs
}

func (pts *ScatterBars) Thumbnail(c *draw.Canvas) {
	c.DrawGlyph(pts.GlyphStyle, c.Center())
}
