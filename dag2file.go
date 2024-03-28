package merkledag

import (
	"encoding/json"
	"strings"
)

const COUNTER = 4

// Hash to file
// Hash2File 函数接收一个KVStore对象、一个hash值、一个路径和一个HashPool对象，
// 根据给定的hash和路径，返回对应的文件。如果hash存在于store中，则获取对应的对象并转换为Object类型，
// 然后根据路径获取文件
func Hash2File(store KVStore, hash []byte, path string, hp HashPool) []byte {
	// 根据hash和path， 返回对应的文件, hash对应的类型是tree
	flag, _ := store.Has(hash)
	if flag {
		objBinary, _ := store.Get(hash)
		obj := binaryToObj(objBinary)
		pathArr := strings.Split(path, "/")
		cur := 1
		return getFileByDir(obj, pathArr, cur, store)
	}
	return nil
}

// getFileByDir 函数接收一个Object对象、一个路径数组、一个当前索引和一个KVStore对象，
// 根据给定的路径和当前索引，从Object对象中获取文件。
func getFileByDir(obj *Object, pathArr []string, cur int, store KVStore) []byte {
	if cur >= len(pathArr) {
		return nil
	}
	index := 0
	for i := range obj.Links {
		objType := string(obj.Data[index : index+COUNTER])
		index += COUNTER
		objInfo := obj.Links[i]
		if objInfo.Name != pathArr[cur] {
			continue
		}
		switch objType {
		case TREE:
			objDirBinary, _ := store.Get(objInfo.Hash)
			objDir := binaryToObj(objDirBinary)
			ans := getFileByDir(objDir, pathArr, cur+1, store)
			if ans != nil {
				return ans
			}
		case BLOB:
			ans, _ := store.Get(objInfo.Hash)
			return ans
		case LIST:
			objLinkBinary, _ := store.Get(objInfo.Hash)
			objList := binaryToObj(objLinkBinary)
			ans := getFileByList(objList, store)
			return ans
		}
	}
	return nil
}

// getFileByList 函数接收一个Object对象和一个KVStore对象，
// 从Object对象中获取文件列表
func getFileByList(obj *Object, store KVStore) []byte {
	ans := make([]byte, 0)
	index := 0
	for i := range obj.Links {
		curObjType := string(obj.Data[index : index+COUNTER])
		index += COUNTER
		curObjLink := obj.Links[i]
		curObjBinary, _ := store.Get(curObjLink.Hash)
		curObj := binaryToObj(curObjBinary)
		// if the type is blob
		if curObjType == BLOB {
			ans = append(ans, curObjBinary...)
		} else { //List
			tmp := getFileByList(curObj, store)
			ans = append(ans, tmp...)
		}
	}
	return ans
}

// binaryToObj 函数接收一个二进制对象，
// 将二进制对象解析为Object对象
func binaryToObj(objBinary []byte) *Object {
	var res Object
	json.Unmarshal(objBinary, &res)
	return &res
}
