package main

func isStrArr(possibleString interface{}) bool {
	defer func  () {
		err := recover()
		if err != nil {
			return false;
		}
	}
	 _ := []string(possibleString)
	 return true;
}

func isStr(possibleString interface{}) bool {
	defer func  () {
		err := recover()
		if err != nil {
			return false;
		}
	}
	 _ := string(possibleString)
	 return true;
}
