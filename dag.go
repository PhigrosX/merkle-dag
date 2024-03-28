package merkledag

import (
	"encoding/json"
	"hash"
)

type Link struct {
	Name string
	Hash []byte
	Size int
}

type Object struct {
	Links []Link
	Data  []byte
}

// dfsForSliceFile 函数用于切分文件对象并递归构建 MerkleDAG
func dfsForSliceFile(hight int, node File, store KVStore, seedId int, h hash.Hash) (*Object, int) {
	if hight == 1 { // 如果高度为1，则切分文件并创建 Blob 对象
		if (len(node.Bytes()) - seedId) <= 256*1024 {
			data := node.Bytes()[seedId:]
			blob := Object{
				Links: nil,
				Data:  data,
			}
			jsonMarshal, _ := json.Marshal(blob)
			h.Reset()
			h.Write(jsonMarshal)
			flag, _ := store.Has(h.Sum(nil))
			if !flag {
				store.Put(h.Sum(nil), data)
			}
			return &blob, len(data)
		}
		links := &Object{}
		lenData := 0
		for i := 1; i <= 4096; i++ {
			end := seedId + 256*1024
			if len(node.Bytes()) < end {
				end = len(node.Bytes())
			}
			data := node.Bytes()[seedId:end]
			blob := Object{
				Links: nil,
				Data:  data,
			}
			lenData += len(data)
			jsonMarshal, _ := json.Marshal(blob)
			h.Reset()
			h.Write(jsonMarshal)
			flag, _ := store.Has(h.Sum(nil))
			if !flag {
				store.Put(h.Sum(nil), data)
			}
			links.Links = append(links.Links, Link{
				Hash: h.Sum(nil),
				Size: len(data),
			})
			links.Data = append(links.Data, []byte("blob")...)
			seedId += 256 * 1024
			if seedId >= len(node.Bytes()) {
				break
			}
		}
		jsonMarshal, _ := json.Marshal(links)
		h.Reset()
		h.Write(jsonMarshal)
		flag, _ := store.Has(h.Sum(nil))
		if !flag {
			store.Put(h.Sum(nil), jsonMarshal)
		}
		return links, lenData
	} else { //// 如果高度大于1，则切分文件并创建链接对象
		links := &Object{}
		lenData := 0
		for i := 1; i <= 4096; i++ {
			if seedId >= len(node.Bytes()) {
				break
			}
			tmp, lens := dfsForSliceFile(hight-1, node, store, seedId, h)
			lenData += lens
			jsonMarshal, _ := json.Marshal(tmp)
			h.Reset()
			h.Write(jsonMarshal)
			links.Links = append(links.Links, Link{
				Hash: h.Sum(nil),
				Size: lens,
			})
			typeName := "link"
			if tmp.Links == nil {
				typeName = "blob"
			}
			links.Data = append(links.Data, []byte(typeName)...)
		}
		jsonMarshal, _ := json.Marshal(links)
		h.Reset()
		h.Write(jsonMarshal)
		flag, _ := store.Has(h.Sum(nil))
		if !flag {
			store.Put(h.Sum(nil), jsonMarshal)
		}
		return links, lenData
	}
}

// sliceFile 函数用于切分文件并创建 MerkleDAG 对象
func sliceFile(node File, store KVStore, h hash.Hash) *Object {
	if len(node.Bytes()) <= 256*1024 {
		data := node.Bytes()
		blob := Object{
			Links: nil,
			Data:  data,
		}
		jsonMarshal, _ := json.Marshal(blob)
		h.Reset()
		h.Write(jsonMarshal)
		flag, _ := store.Has(h.Sum(nil))
		if !flag {
			store.Put(h.Sum(nil), data)
		}
		return &blob
	}
	linkLen := (len(node.Bytes()) + (256*1024 - 1)) / (256 * 1024)
	hight := 0
	tmp := linkLen
	for {
		hight++
		tmp /= 4096
		// fmt.Println(tmp)
		if tmp == 0 {
			break
		}
	}
	// fmt.Println(hight)
	res, _ := dfsForSliceFile(hight, node, store, 0, h)
	return res
}

// sliceDir 函数用于切分目录并创建 MerkleDAG 对象
func sliceDir(node Dir, store KVStore, h hash.Hash) *Object {
	iter := node.It()
	treeObject := &Object{}
	for iter.Next() {
		node := iter.Node()
		if node.Type() == FILE {
			file := node.(File)
			tmp := sliceFile(file, store, h)
			jsonMarshal, _ := json.Marshal(tmp)
			h.Reset()
			h.Write(jsonMarshal)
			treeObject.Links = append(treeObject.Links, Link{
				Hash: h.Sum(nil),
				Size: int(file.Size()),
				Name: file.Name(),
			})
			typeName := "link"
			if tmp.Links == nil {
				typeName = "blob"
			}
			treeObject.Data = append(treeObject.Data, []byte(typeName)...)
		} else {
			dir := node.(Dir)
			tmp := sliceDir(dir, store, h)
			jsonMarshal, _ := json.Marshal(tmp)
			h.Reset()
			h.Write(jsonMarshal)
			treeObject.Links = append(treeObject.Links, Link{
				Hash: h.Sum(nil),
				Size: int(dir.Size()),
				Name: dir.Name(),
			})
			typeName := "tree"
			treeObject.Data = append(treeObject.Data, []byte(typeName)...)
		}
	}
	jsonMarshal, _ := json.Marshal(treeObject)
	h.Reset()
	h.Write(jsonMarshal)
	flag, _ := store.Has(h.Sum(nil))
	if !flag {
		store.Put(h.Sum(nil), jsonMarshal)
	}
	return treeObject
}

func Add(store KVStore, node Node, h hash.Hash) []byte {
	// TODO 将分片写入到KVStore中，并返回Merkle Root
	if node.Type() == FILE {
		file := node.(File)
		tmp := sliceFile(file, store, h)
		jsonMarshal, _ := json.Marshal(tmp)
		h.Write(jsonMarshal)
		return h.Sum(nil)
	} else {
		dir := node.(Dir)
		tmp := sliceDir(dir, store, h)
		jsonMarshal, _ := json.Marshal(tmp)
		h.Write(jsonMarshal)
		return h.Sum(nil)
	}
}
