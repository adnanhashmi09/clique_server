package utils

import "github.com/gocql/gocql"

func RemoveElementFromArray(arr []string, element string) []string {
	var result []string

	for _, value := range arr {
		if value != element {
			result = append(result, value)
		}
	}

	return result
}

func RemoveElementFromArrayUUID(arr []gocql.UUID, element gocql.UUID) []gocql.UUID {
	var result []gocql.UUID

	for _, value := range arr {
		if value != element {
			result = append(result, value)
		}
	}

	return result
}
