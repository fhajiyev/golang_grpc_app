package service

import "strconv"

// AddItemToList func definition
func AddItemToList(list []string, itemID int, isAdd bool) []string {
	if len(list) == 0 {
		if isAdd {
			list = append(list, strconv.Itoa(itemID))
		}
	} else {
	loop:
		for idx, idStr := range list {
			id, _ := strconv.Atoi(idStr)
			if isAdd {
				switch {
				case id == itemID:
					break loop
				case id > itemID:
					list = append(list[:idx], append([]string{strconv.Itoa(itemID)}, list[idx:]...)...)
					break loop
				case idx == len(list)-1:
					list = append(list, strconv.Itoa(itemID))
					break loop
				}
			} else {
				if id == itemID {
					list = append(list[:idx], list[idx+1:]...)
					break loop
				}
			}
		}
	}

	return list
}
