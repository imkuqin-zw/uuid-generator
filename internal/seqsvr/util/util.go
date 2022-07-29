package util

import (
	"github.com/imkuqin-zw/uuid-generator/internal/seqsvr/proto"
)

// const SectionSize = 100000

// 计算section的ID
func CalcSectionIDByID(id uint32, sectionSize uint32) uint32 {
	return id/sectionSize*sectionSize + 1
}

// 计算序列号在section中的位置
func CalcSeqIdxByID(id uint32, sectionSize uint32) uint32 {
	return (id - 1) % sectionSize
}

func MemSet(a []uint64, v uint64) {
	if len(a) == 0 {
		return
	}
	a[0] = v
	for bp := 1; bp < len(a); bp *= 2 {
		copy(a[bp:], a[:bp])
	}
}

func MakeSeq(val uint64, sectionSize uint32) []uint64 {
	seq := make([]uint64, sectionSize)
	MemSet(seq, val)
	return seq
}

func SearchSetByID(setRanges [][]uint32, id uint32) (int, []uint32) {
	i, j := 0, len(setRanges)
	for i <= j {
		h := int(uint(i+j) >> 1)
		setRange := setRanges[h]
		if setRange[0] <= id && id < setRange[0]+setRange[1] {
			return h, setRange
		} else if setRange[0]+setRange[1] <= id {
			i = h + 1
		} else {
			j = h - 1
		}
	}
	return -1, nil
}

func HasSection(sections []uint32, sectionID uint32) bool {
	i, j := 0, len(sections)
	for i <= j {
		h := int(uint(i+j) >> 1)
		if sections[h] == sectionID {
			return true
		} else if sections[h] < sectionID {
			i = h + 1
		} else {
			j = h - 1
		}
	}
	return false
}

func SearchSet(sets []*proto.Set, setID uint32) *proto.Set {
	i, j := 0, len(sets)
	for i <= j {
		h := int(uint(i+j) >> 1)
		set := sets[h]
		if set.ID == setID {
			return set
		} else if set.ID < setID {
			i = h + 1
		} else {
			j = h - 1
		}
	}
	return nil
}

func SearchNode(nodes []*proto.AllocNode, IP string) *proto.AllocNode {
	i, j := 0, len(nodes)
	for i <= j {
		h := int(uint(i+j) >> 1)
		node := nodes[h]
		if node.IP == IP {
			return node
		} else if node.IP < IP {
			i = h + 1
		} else {
			j = h - 1
		}
	}
	return nil
}
