package merkledag

import (
	"encoding/json"
	"strings"
)

// 定义步长为4
const STEP = 4

// Hash to file
// Hash2File 函数根据给定的哈希和路径，从KVStore中获取对应的文件
func Hash2File(store KVStore, hash []byte, path string, hp HashPool) []byte {
	// 根据hash和path， 返回对应的文件, hash对应的类型是tree
	flag, _ := store.Has(hash) // 检查哈希是否存在于存储中
	if flag {
		objBinary, _ := store.Get(hash)
		obj := binaryToObj(objBinary)
		pathArr := strings.Split(path, "\\") // 将路径分割为数组
		cur := 1
		return getFileByDir(obj, pathArr, cur, store) // 根据目录获取文件
	}
	return nil
}

// getFileByDir 函数根据给定的目录和路径，从KVStore中获取对应的文件
func getFileByDir(obj *Object, pathArr []string, cur int, store KVStore) []byte {
	if cur >= len(pathArr) {
		return nil
	}
	index := 0
	for i := range obj.Links {
		objType := string(obj.Data[index : index+STEP])
		index += STEP
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

// getFileByList 函数根据给定的列表和KVStore，获取对应的文件
func getFileByList(obj *Object, store KVStore) []byte {
	ans := make([]byte, 0)
	index := 0
	for i := range obj.Links {
		curObjType := string(obj.Data[index : index+STEP])
		index += STEP
		curObjLink := obj.Links[i]
		curObjBinary, _ := store.Get(curObjLink.Hash)
		curObj := binaryToObj(curObjBinary)
		if curObjType == BLOB {
			ans = append(ans, curObjBinary...)
		} else { //List
			tmp := getFileByList(curObj, store)
			ans = append(ans, tmp...)
		}
	}
	return ans
}

// binaryToObj 函数将二进制数据转换为对象
func binaryToObj(objBinary []byte) *Object {
	var res Object
	json.Unmarshal(objBinary, &res) // 使用json.Unmarshal函数将二进制数据转换为对象
	return &res
}
