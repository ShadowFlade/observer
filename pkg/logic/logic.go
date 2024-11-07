package logic

import (
	"fmt"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/ShadowFlade/observer/pkg/db"
	"github.com/charmbracelet/lipgloss"
	"github.com/jmoiron/sqlx"
)

type DEBUG_STATE int64

var (
	logoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#01FAC6")).Bold(true)
	usersStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("190")).Italic(true).Width(8)
	memStyle   = lipgloss.NewStyle().PaddingLeft(1).Bold(true).Align(lipgloss.Right)
)

const (
	DEBUG_NONE DEBUG_STATE = iota
	DEBUG_INFO
	DEBUG_ERROR
	DEBUG_DEBUG
)

type App struct {
	DebugState DEBUG_STATE
	DB         sqlx.DB
}
type TopColumns struct {
	PID, User, PR, NI, Virt, Res, SHR, S, CPU, Mem, Time, Prog int
}

type ProgStat struct {
	Name     string
	MemUsage float32
}

type UserStat struct {
	Prog                 []ProgStat
	TotalMemUsage        float32
	TotalMemUsagePercent float32
	Name                 userName
}

type userName string
type UserStats map[userName]UserStat

func (a *App) Main(onlyUser userName, intervalSeconds int, db db.Db) {
	regularUsers, ids := db.GetRegularUsers()
	onlyUser = a.formatUsernameTop(onlyUser)
	interval := intervalSeconds * int(time.Second)
	ticker := time.NewTicker(time.Duration(interval))
	defer ticker.Stop()
	done := make(chan bool)
	go func() {
		users, userStats, _ := a.parseTopAndGetUserResults(onlyUser)

		for _, user := range users {
			fmt.Printf(
				"%s %s\n",
				usersStyle.Render(string(user)),
				memStyle.Render(
					strconv.FormatFloat(float64(userStats[user].TotalMemUsage), 'f', 2, 32)),
			)
		}
		for i, user := range users {
			if !slices.Contains(users, user) {
				isOk := a.checkWriteRegularUser(user)
				if isOk {
					regularUsers = append(regularUsers, string(user))
					db.WriteStats(userStats[user].TotalMemUsage, userStats[user].TotalMemUsagePercent, ids[i], len(users))
				}

			} else {
				db.WriteStats(userStats[user].TotalMemUsage, userStats[user].TotalMemUsagePercent, ids[i], len(users))
			}
		}
	}()

	done <- true
	fmt.Println("Ticker done")

}

func (a *App) parseTopAndGetUserResults(onlyUser userName) ([]userName, UserStats, float32) {

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

	userStats := make(map[userName]UserStat)
	var totalEmployeesMemoryUsage float32

	if err != nil {
		fmt.Println("Error outputting 'top'")
	}

	users := []userName{}

	for _, val := range strings.Split(string(topStats), "\n") {
		splitStr := strings.Split(val, " ")

		if len(splitStr) < 3 {
			continue
		}

		user, mem, prog := userName(strings.Trim(splitStr[0], " ")), splitStr[1], splitStr[2]

		//later can refocator that if we pass onluUser we can sort columhns beforehande - so before topStatsCommand so we get only one user - we sort on user, and after stopped getting these users commands we stop parsing

		if onlyUser != "" && user != onlyUser {
			continue
		}

		memInt, err := strconv.Atoi(mem)
		totalEmployeesMemoryUsage += float32(memInt)

		if userStat, ok := userStats[user]; ok && !slices.Contains(users, user) {
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
	for _, userStat := range userStats {
		userStat.TotalMemUsagePercent = userStat.TotalMemUsage / totalEmployeesMemoryUsage
	}
	fmt.Println(len(users), " length of users")
	return users, userStats, totalEmployeesMemoryUsage
}

func (a *App) formatUsernameTop(username userName) userName {
	count := utf8.RuneCountInString(string(username))
	if count > 7 {
		return username[:7] + "+"
	} else {
		return username
	}
}

func (a *App) checkWriteRegularUser(user userName) bool {
	db := db.Db{}
	db.Connect()
	command := "less /etc/passwd"
	cmd := exec.Command("bash", "-c", command)

	users, err := cmd.Output()

	if err != nil {
		panic("Cannot write regular users")
	}

	r, _ := regexp.Compile(fmt.Sprintf(`%s\:x\:(\d+).*`, user))
	res := r.Find(users) //we dont count users with groupid less than 1000 bc its system users
	if int(res[2]) > 1000 {
		db.WriteRegularUser(string(res[1]))
		return true
	}
	return false
}
