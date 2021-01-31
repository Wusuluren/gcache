package gcache

const defaultSlotNum = 0xFF

func hashInt(v int) int64 {
	return int64(v) & defaultSlotNum
}

func hashString(v string) int64 {
	sum := 0
	for i := range v {
		sum += int(v[i])
	}
	return hashInt(sum)
}
