package distance

func Levenshtein(first string, second string) int {
	firstRunes := []rune(first)
	secondRunes := []rune(second)

	firstLen := len(firstRunes)
	secondLen := len(secondRunes)

	if firstLen == 0 {
		return secondLen
	}
	if secondLen == 0 {
		return firstLen
	}

	dp := make([][]int, firstLen+1)
	for i := range dp {
		dp[i] = make([]int, secondLen+1)
	}

	for i := 0; i <= firstLen; i++ {
		dp[i][0] = i
	}
	for j := 0; j <= secondLen; j++ {
		dp[0][j] = j
	}

	for i := 1; i <= firstLen; i++ {
		for j := 1; j <= secondLen; j++ {
			cost := 1
			if firstRunes[i-1] == secondRunes[j-1] {
				cost = 0
			}

			deleteCost := dp[i-1][j] + 1    // delete rune from first string
			insertCost := dp[i][j-1] + 1    // insert rune into first string
			replaceCost := dp[i-1][j-1] + cost // replace (or match when cost=0)

			dp[i][j] = min3(deleteCost, insertCost, replaceCost)
		}
	}

	return dp[firstLen][secondLen]
}

func min3(first int, second int, third int) int {
	minimum := first
	if second < minimum {
		minimum = second
	}
	if third < minimum {
		minimum = third
	}
	return minimum
}
