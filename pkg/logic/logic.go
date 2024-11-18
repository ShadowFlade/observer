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
	PID, UserName, PR, NI, Virt, Res, SHR, S, CPU, Mem, Time, Prog int
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

	for _, user := range lessUsers {
		fUser := this.FormatUsernameTop(user.UserName)

		if !this.IsRegularUser(user.UserName, user.Id) {
			continue
		}

		renderer.RenderUser(string(user.UserName), userStats[fUser].TotalMemUsage)
		fmt.Printf("%v\n", userStats["shadowf+"])

		if !slices.Contains(regularUsers, string(user.UserName)) {
			isOk := this.checkWriteRegularUser(UserName(user.UserName), db)
			if isOk {
				regularUsers = append(regularUsers, user.UserName)
				db.WriteStats(
					userStats[UserName(user.UserName)].TotalMemUsage,
					userStats[UserName(user.UserName)].TotalMemUsagePercent,
					user.Id,
					len(lessUsers),
				)
			}

		} else {
			db.WriteStats(
				userStats[UserName(user.UserName)].TotalMemUsage,
				userStats[UserName(user.UserName)].TotalMemUsagePercent,
				user.Id,
				len(lessUsers),
			)
		}
	}

}

func (this *App) parseTopAndGetUserResults(onlyUser UserName) (UserStats, float32) {

	topColumns := TopColumns{PID: 1, UserName: 2, PR: 3, NI: 4, Virt: 5, Res: 6, SHR: 7, S: 8, CPU: 9, Mem: 10, Time: 11, Prog: 12}
	isSkipHeader := true
	var header string

	if isSkipHeader {
		header = "NR>7"
	} else {
		header = ""
	}

	topStatsCommand := fmt.Sprintf("top -b -n 1 | awk '%s {print $%d, $%d, $%d}'", header, topColumns.UserName, topColumns.Res, topColumns.Prog)

	cmd := exec.Command("bash", "-c", topStatsCommand)
	topStats, err := cmd.Output()

	userStats := make(map[UserName]UserStat)
	var totalEmployeesMemoryUsage float32

	if err != nil {
		panic("Error outputting 'top'")
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
				log.Printf("Can't convert user memory usage into int %v", mem)
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

func (this *App) checkWriteRegularUser(userName UserName, db db.Db) bool {
	command := "less /etc/passwd"
	cmd := exec.Command("bash", "-c", command)

	users, err := cmd.Output()

	if err != nil {
		panic("Cannot write regular users")
	}

	if userName == "" {
		userName = ".*"
	}

	fmt.Println(userName, " username")
	regex := fmt.Sprintf(`(%s)\:x\:(\d+).*`, userName)
	r, _ := regexp.Compile(regex)
	res := r.FindSubmatch(users)

	foundId, err := strconv.Atoi(string(res[2]))

	if err != nil {
		panic("Could not find user id")
	}

	if this.IsRegularUser(string(userName), foundId) {
		db.WriteRegularUser(string(res[1]), int32(foundId))
		return true
	}
	return false
}

type UserAndId struct {
	UserName string
	Id       int
}

func (this *App) GetUsers() []UserAndId {

	command := "less /etc/passwd"
	cmd := exec.Command("bash", "-c", command)

	usersTop, err := cmd.Output()

	if err != nil {
		panic("Cannot write regular users")
	}

	r, _ := regexp.Compile(`(.*)\:x\:(\d+).*`)
	res := r.FindAllSubmatch(usersTop, -1) //we dont count users with groupid less than 1000 bc its system users
	var users []UserAndId

	for _, tuple := range res {
		// if index == 0 {
		// 	fmt.Println(string(tuple[0]), string(tuple[1]), string(tuple[2]), " tuple")
		// }
		id, err := strconv.Atoi(string(tuple[2]))

		if err != nil {
			log.Fatal(err, string(tuple[2]))
		}

		users = append(
			users,
			UserAndId{
				UserName: string(tuple[1]),
				Id:       id,
			},
		)
	}

	return users
}

func (this *App) IsRegularUser(userName string, id int) bool {
	return id >= 1000 && userName != "nobody"
}
