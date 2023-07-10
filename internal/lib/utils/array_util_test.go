/*
 * @Descripttion:
 * @version:
 * @Date: 2023-05-24 14:16:01
 * @LastEditTime: 2023-05-24 14:17:59
 */
package utils

import (
	"fmt"
	"testing"
)

func TestReverse(t *testing.T) {
	arr := []string{"1", "2", "3"}
	Reverse(arr)
	fmt.Println(arr)
}
