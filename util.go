package gcache

const defaultSlotNum = 0xFF

func hashInt(v int) int {
	return v & defaultSlotNum
}

func hashInt64(v int64) int {
	return int(v & defaultSlotNum)
}

func hashString(v string) int {
	sum := 0
	for i := range v {
		sum += int(v[i])
	}
	return hashInt(sum)
}
