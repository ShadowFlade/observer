package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	topStatsCommand := "top -b -n 1 | awk 'NR>7 {print $2, $6, $12}'"
	cmd := exec.Command("bash", "-c", topStatsCommand)
	topStats, err := cmd.Output()

	type ProgStat struct {
		Name     string
		MemUsage float32
	}
	type UserStat struct {
		Prog          []ProgStat
		TotalMemUsage float32
		Name          string
	}

	userStats := make(map[string]UserStat)

	if err != nil {
		fmt.Println("Error outputting 'top'")
	}

	for _, val := range strings.Split(string(topStats), "\n") {
		splitStr := strings.Split(val, " ")

		if len(splitStr) < 3 {
			continue
		}

		user, mem, prog := splitStr[0], splitStr[1], splitStr[2]

		memInt, err := strconv.Atoi(mem)
		if userStat, ok := userStats[user]; ok {

			if err != nil {
				fmt.Println("Can't convert user memory usage into int", " ", mem)
			}

			userStat.TotalMemUsage += float32(memInt)
			userStat.Prog = append(userStat.Prog, ProgStat{
				Name:     prog,
				MemUsage: float32(memInt),
			})
			userStats[user] = userStat
		} else {
			userStats[user] = UserStat{
				Prog: []ProgStat{
					{
						Name:     prog,
						MemUsage: float32(memInt),
					},
				},
				Name:          user,
				TotalMemUsage: float32(memInt),
			}
		}
	}
	fmt.Println(userStats, " user stats")
}
