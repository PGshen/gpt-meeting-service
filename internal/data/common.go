/*
 * @Descripttion:
 * @version:
 * @Date: 2023-05-03 15:24:55
 * @LastEditTime: 2023-05-03 15:26:07
 */
package data

func paginate(pageNum, pageSize int64) (int64, int64) {
	if pageNum == 0 {
		pageNum = 1
	}
	switch {
	case pageSize > 100:
		pageSize = 100
	case pageSize < 0:
		pageSize = 10
	}
	offset := (pageNum - 1) * pageSize
	return offset, pageSize
}
