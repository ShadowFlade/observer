package logic

import (
	"fmt"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

type DEBUG_STATE int64

const (
	DEBUG_NONE DEBUG_STATE = iota
	DEBUG_INFO
	DEBUG_ERROR
	DEBUG_DEBUG
)

type App struct {
	DebugState DEBUG_STATE
}
type TopColumns struct {
	PID, User, PR, NI, Virt, Res, SHR, S, CPU, Mem, Time, Prog int
}

type ProgStat struct {
	Name     string
	MemUsage float32
}

type UserStat struct {
	Prog          []ProgStat
	TotalMemUsage float32
	Name          string
}

type UserStats map[string]UserStat

func (a *App) Main(onlyUser string) ([]string, UserStats) {
	onlyUser = a.formatUsernameTop(onlyUser)
	topColumns := TopColumns{PID: 1, User: 2, PR: 3, NI: 4, Virt: 5, Res: 6, SHR: 7, S: 8, CPU: 9, Mem: 10, Time: 11, Prog: 12}
	isSkipHeader := true
	var header string

	if isSkipHeader {
		header = "NR>7"
	} else {
		header = ""
	}

	topStatsCommand := fmt.Sprintf("top -b -n 1 | awk '%s {print $%d, $%d, $%d}'", header, topColumns.User, topColumns.Res, topColumns.Prog)

	cmd := exec.Command("bash", "-c", topStatsCommand)
	topStats, err := cmd.Output()

	userStats := make(map[string]UserStat)

	if err != nil {
		fmt.Println("Error outputting 'top'")
	}

	users := []string{}

	for _, val := range strings.Split(string(topStats), "\n") {
		splitStr := strings.Split(val, " ")

		if len(splitStr) < 3 {
			continue
		}
		user, mem, prog := strings.Trim(splitStr[0]," "), splitStr[1], splitStr[2]

		//later can refocator that if we pass onluUser we can sort columhns beforehande - so before topStatsCommand so we get only one user - we sort on user, and after stopped getting these users commands we stop parsing

		if onlyUser != "" && user != onlyUser {
			continue
		}

		memInt, err := strconv.Atoi(mem)

		if userStat, ok := userStats[user]; ok && !slices.Contains(users,user) {
			users = append(users, user)

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
	fmt.Println(len(users)," length of users")
	return users, userStats
}

func (a *App) formatUsernameTop(username string) string {
	count := utf8.RuneCountInString(username)
	if  count > 7 {
		return username[:7] + "+"
	} else {
		return username
	}
}
