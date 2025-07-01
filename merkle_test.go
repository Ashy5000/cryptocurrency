package main

import (
	"fmt"
	"testing"

	. "cryptocurrency/node_util"

	"github.com/stretchr/testify/assert"
)

func TestMerkleNode(t *testing.T) {
	node := MerkleNode{
		Data:     []byte("test"),
		Children: make(map[byte]int),
		Parent:   0,
		Hash:     "",
		Key:      "",
	}
	assert.NotNil(t, node)
}

func TestHashNode(t *testing.T) {
	tree := []MerkleNode{
		{
			Data:     []byte("test"),
			Children: make(map[byte]int),
			Parent:   0,
			Hash:     "",
			Key:      "",
		},
	}
	hash := HashNode(tree, 0)
	assert.NotNil(t, hash)
}

func TestHashTree(t *testing.T) {
	tree := []MerkleNode{}
	tree = InsertValue(tree, "test", []byte("test"))
	tree = HashTree(tree, 0)
	assert.NotNil(t, tree[0].Hash)
	fmt.Println(tree[0].Hash)
}

func TestInsertAndGetValue(t *testing.T) {
	var tree []MerkleNode
	key := "test"
	value := []byte("test")

	tree = InsertValue(tree, key, value)

	result, found := GetValue(tree, key)

	assert.True(t, found)
	assert.Equal(t, value, result)
}

func TestMerge(t *testing.T) {
	tree := []MerkleNode{
		{
			Data:     []byte("test"),
			Children: make(map[byte]int),
			Parent:   0,
			Hash:     "",
			Key:      "",
		},
	}
	delta := []MerkleNode{
		{
			Data:     []byte("test2"),
			Children: make(map[byte]int),
			Parent:   0,
			Hash:     "",
			Key:      "",
		},
	}
	Merge(tree, delta)
	assert.NotNil(t, tree[0].Data)
	assert.Equal(t, []byte("test"), tree[0].Data)
}
