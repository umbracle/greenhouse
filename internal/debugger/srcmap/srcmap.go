package srcmap

import (
	"fmt"
	"strconv"
	"strings"
)

type SrcMap []*SrcLocation

type SrcLocation struct {
	Index     int
	Offset    int
	Length    int
	FileIndex int
	Jump      string
}

func (s *SrcLocation) Src() string {
	return fmt.Sprintf("%d:%d:%d", s.Offset, s.Length, s.FileIndex)
}

func (s *SrcLocation) String() string {
	return fmt.Sprintf("Index: %d, Offset: %d, Length: %d, FileIndex: %d, Jump: %s", s.Index, s.Offset, s.Length, s.FileIndex, s.Jump)
}

func (s *SrcLocation) Copy() *SrcLocation {
	ss := new(SrcLocation)
	*ss = *s
	return ss
}

func ParseSrcMap(m string) (SrcMap, error) {
	res := SrcMap{}

	parseInt := func(s string) (int, error) {
		return strconv.Atoi(s)
	}

	var err error
	lastEntry := &SrcLocation{}

	for indx, entry := range strings.Split(m, ";") {
		current := strings.Split(entry, ":")

		var num int
		if len(entry) != 0 {
			num = len(current)
		}

		if num >= 1 && current[0] != "" {
			// offset
			if lastEntry.Offset, err = parseInt(current[0]); err != nil {
				return nil, err
			}
		}
		if num >= 2 && current[1] != "" {
			// length
			if lastEntry.Length, err = parseInt(current[1]); err != nil {
				return nil, err
			}
		}
		if num >= 3 && current[2] != "" {
			// file index
			if lastEntry.FileIndex, err = parseInt(current[2]); err != nil {
				return nil, err
			}
		}
		if num >= 4 && current[3] != "" {
			lastEntry.Jump = current[3]
		}

		entry := lastEntry.Copy()
		entry.Index = indx
		res = append(res, entry)
	}
	return res, nil
}
