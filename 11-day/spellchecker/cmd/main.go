package main

func findThePrefixCommonArray(A []int, B []int) []int {
	arr1 := make([]int, len(A)+1)
	arr2 := make([]int, len(B)+1)

	ans := make([]int, len(A)+1)

	cumulative := 0

	for i := 0; i < len(arr1) && i < len(arr2); i++ {
		if arr1[i] == arr2[i] {
			cumulative++
			continue
		}
		arr1[A[i]]++
		arr2[B[i]]++

		if arr1[B[i]] != 0 {
			cumulative++
		}

		if arr2[B[i]] != 0 {
			cumulative++
		}

		ans[i] = cumulative
	}

	return ans
}
