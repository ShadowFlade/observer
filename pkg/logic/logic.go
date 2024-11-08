package logic

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/ShadowFlade/observer/pkg/db"
	"github.com/ShadowFlade/observer/pkg/render"
	"github.com/jmoiron/sqlx"
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
	Name                 UserName
}

type UserName string
type UserStats map[UserName]UserStat

func (this *App) Main(
	onlyUser string,
	intervalSeconds int,
	db db.Db,
	regularUsers []string,
	ids []int,
) {
	formattedUser := this.FormatUsernameTop(onlyUser)
	lessUsers := this.GetUsers()
	userStats, _ := this.parseTopAndGetUserResults(formattedUser)
	renderer := render.Renderer{}

	for i, user := range lessUsers {
		fUser := this.FormatUsernameTop(user.User)
		renderer.RenderUser(string(user.User), userStats[fUser].TotalMemUsage)

		if !slices.Contains(regularUsers, string(user.User)) {
			isOk := this.checkWriteRegularUser(string(user.User), db)
			if isOk {
				regularUsers = append(regularUsers, string(user.User))
				db.WriteStats(
					userStats[UserName(user.User)].TotalMemUsage,
					userStats[UserName(user.User)].TotalMemUsagePercent,
					ids[i],
					len(lessUsers),
				)
			}

		} else {
			db.WriteStats(
				userStats[UserName(user.User)].TotalMemUsage,
				userStats[UserName(user.User)].TotalMemUsagePercent,
				ids[i],
				len(lessUsers),
			)
		}
	}

}

func (this *App) parseTopAndGetUserResults(onlyUser UserName) (UserStats, float32) {

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

	userStats := make(map[UserName]UserStat)
	var totalEmployeesMemoryUsage float32

	if err != nil {
		fmt.Println("Error outputting 'top'")
	}

	users := []UserName{}

	for _, val := range strings.Split(string(topStats), "\n") {
		splitStr := strings.Split(val, " ")

		if len(splitStr) < 3 {
			continue
		}

		user, mem, prog := UserName(strings.Trim(splitStr[0], " ")), splitStr[1], splitStr[2]

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

	return userStats, totalEmployeesMemoryUsage
}

func (this *App) FormatUsernameTop(username string) UserName {
	count := utf8.RuneCountInString(username)
	if count > 7 {
		return UserName(username[:7] + "+")
	} else {
		return UserName(username)
	}
}

func (this *App) checkWriteRegularUser(user string, db db.Db) bool {
	command := "less /etc/passwd"
	cmd := exec.Command("bash", "-c", command)

	users, err := cmd.Output()
	fmt.Println(string(users), " users output", user, ": user")

	if err != nil {
		panic("Cannot write regular users")
	}

	if user == "" {
		user = ".*"
	}
	r, _ := regexp.Compile(fmt.Sprintf(`(%s)\:x\:(\d+).*`, user))
	res := r.Find(users) //we dont count users with groupid less than 1000 bc its system users
	fmt.Println(res)
	if int(res[2]) > 1000 {
		db.WriteRegularUser(string(res[1]))
		return true
	}
	return false
}

type UserAndId struct {
	User string
	Id   int
}

func (this *App) GetUsers() []UserAndId {

	command := "less /etc/passwd"
	cmd := exec.Command("bash", "-c", command)

	usersTop, err := cmd.Output()

	if err != nil {
		panic("Cannot write regular users")
	}

	r, _ := regexp.Compile(`(.*)\:x\:(\d+).*`)
	res := r.FindAll(usersTop, -1) //we dont count users with groupid less than 1000 bc its system users
	var users []UserAndId

	for _, tuple := range res {
		id, err := strconv.Atoi(string(tuple[1]))

		if err != nil {
			log.Fatal(err, string(tuple))
		}

		users = append(
			users,
			UserAndId{
				User: string(tuple[0]),
				Id:   id,
			},
		)
	}
	fmt.Println(res, " ITS A RES")

	return []UserAndId{}
}
