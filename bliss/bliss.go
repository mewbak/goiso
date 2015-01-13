package bliss

/*
  Copyright (c) 2014 Tim Henderson
  Released under the GNU General Public License version 3.
  
  This file is part of goiso a wrapper around bliss.
  
  bliss is free software: you can redistribute it and/or modify
  it under the terms of the GNU General Public License version 3
  as published by the Free Software Foundation.
  
  bliss is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.
  
  You should have received a copy of the GNU General Public License
  along with Foobar.  If not, see <http://www.gnu.org/licenses/>.
*/

/*
#cgo LDFLAGS: -lstdc++
#include <stdio.h>
#include "bliss_C.h"

void
hook(void *user_param, unsigned int N, const unsigned int *aut) {
	return;
}
*/
import "C"

import "unsafe"

type BlissGraph C.struct_bliss_graph_struct

// A context manager which release the graph after the block
// ends.
func Graph(nodes int, block func(*BlissGraph)) {
	g := NewGraph(nodes)
	defer g.Release()
	block(g)
}

// Constructs a new bliss graph object. Note, this
// is a directed graph.
// nodes = the number of nodes to add with color 0
func NewGraph(nodes int) *BlissGraph {
	n := C.uint(uint(nodes))
	return (*BlissGraph)(C.bliss_new(n))
}

// Release the graph. You need to manage the memory manually
// as the graph lives in C land.
func (g *BlissGraph) Release() {
	G := (*C.struct_bliss_graph_struct)(g)
	C.bliss_release(G)
}

// Add a new vertex of the given color to the graph.
// The vertex id will be returned.
func (g *BlissGraph) AddVertex(color uint) uint {
	c := C.uint(color)
	G := (*C.struct_bliss_graph_struct)(g)
	return uint(C.bliss_add_vertex(G, c))
}

// Add a new edge between the two vertex ids
// Since this is a directed graph: a -> b
func (g *BlissGraph) AddEdge(a, b uint) {
	x := C.uint(a)
	y := C.uint(b)
	G := (*C.struct_bliss_graph_struct)(g)
	C.bliss_add_edge(G, x, y)
}

// Compare two graphs.
// If a < b: -1
// If a = b: 0
// If a > b: 1
// Note, this does not check for isomorphism. To
// do that either:
//
//     a.Canonical().Cmp(b.Canonical())
//
//  or
//
//     a.Iso(b)
//
// If you are planning the compute and store the canonical
// labeling then the first option is better.
func (a *BlissGraph) Cmp(b *BlissGraph) int {
	g1 := (*C.struct_bliss_graph_struct)(a)
	g2 := (*C.struct_bliss_graph_struct)(b)
	return int(C.bliss_cmp(g1, g2))
}

// Are the graphs isomorphic? This function will compute
// the canonical labeling anew each time. You can compute
// the labeling once and use the Cmp function to compare the
// graphs instead to save time.
func (a *BlissGraph) Iso(b *BlissGraph) bool {
	var cmp bool
	a.CanonicalCtx(func(g1 *BlissGraph) {
		b.CanonicalCtx(func(g2 *BlissGraph) {
			cmp = g1.Cmp(g2) == 0
		})
	})
	return cmp
}

// Compute the canonical labeling. This will function will
// return a new *BlissGraph. If you no longer need the
// original be sure to release it with. Release(). When
// you are done with the canonical graph be sure to release
// it as well.
func (g *BlissGraph) Canonical() *BlissGraph {
	var stats C.struct_bliss_stats_struct
	G := (*C.struct_bliss_graph_struct)(g)
	p := C.bliss_find_canonical_labeling(G, (*[0]byte)(C.hook), nil, &stats)
	return (*BlissGraph)(C.bliss_permute(G, p))
}

// A context manager for Canonical graph.
func (g *BlissGraph) CanonicalCtx(block func(*BlissGraph)) {
	can := g.Canonical()
	defer can.Release()
	block(can)
}

// Compute the permutation. Returns a slice of new indexes. Read the slice as:
// mapping[original-index-for-v] -> new-index-for-v
// If you want to preserve the orginal vertex id's or know how the canonical
// labeling actually maps to the original graph you need to use this method.
func (g *BlissGraph) CanonicalPermutation() (mapping []uint) {
	var stats C.struct_bliss_stats_struct
	G := (*C.struct_bliss_graph_struct)(g)
	N := uint(C.bliss_get_nof_vertices(G))
	p := C.bliss_find_canonical_labeling(G, (*[0]byte)(C.hook), nil, &stats)
	mapping = make([]uint, 0, N)
	for i := uint(0); i < N; i++ {
		ptr := (uintptr(unsafe.Pointer(p)) + uintptr(i)*4)
		var idx uint = uint(*(*C.uint)(unsafe.Pointer(ptr)))
		mapping = append(mapping, idx)
	}
	return mapping
}

