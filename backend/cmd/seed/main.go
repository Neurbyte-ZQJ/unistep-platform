// 测试数据生成脚本（中等规模，覆盖全部业务模块）
//
// 用法：
//
//	cd backend
//	$env:PATH = "D:\msys64\mingw64\bin;" + $env:PATH   # 需要 CGO
//	go run ./cmd/seed                                  # 写入默认 data/unistep.db
//	go run ./cmd/seed -reset                           # 先清空业务表再重新生成
//	go run ./cmd/seed -db ./data/unistep.db -reset     # 指定数据库路径
//
// 生成内容（中等规模）：
//
//	用户：admin/teacher/学生干部/学生 共约 50 人（含已存在的种子账号）
//	团员档案：50 份，覆盖 5 个发展阶段，含申请/积极分子/发展对象/政审记录
//	社团活动：20 个，覆盖 8 种状态，含报名与签到
//	社区队伍：10 支（自治/志愿/值班），含成员、值班排班与志愿服务
//	勤工助学：20 个岗位，含报名、考勤与薪资
package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"unistep-platform/backend/internal/database"
	"unistep-platform/backend/internal/models"
)

// ---- 字典数据 ----

var (
	surnames = []string{"王", "李", "张", "刘", "陈", "杨", "赵", "黄", "周", "吴", "徐", "孙", "马", "朱", "胡", "郭", "何", "高", "林", "罗"}
	names    = []string{"伟", "芳", "娜", "敏", "静", "丽", "强", "磊", "军", "洋", "勇", "艳", "杰", "娟", "涛", "明", "超", "秀", "霞", "平", "刚", "桂英"}
	colleges = []string{"计算机学院", "电子信息学院", "机械工程学院", "经济管理学院", "外国语学院", "马克思主义学院"}
	majors   = map[string][]string{
		"计算机学院":   {"计算机科学与技术", "软件工程", "人工智能", "数据科学"},
		"电子信息学院":  {"通信工程", "电子信息工程", "微电子"},
		"机械工程学院":  {"机械设计制造", "车辆工程", "智能制造"},
		"经济管理学院":  {"会计学", "工商管理", "金融学"},
		"外国语学院":   {"英语", "日语", "翻译"},
		"马克思主义学院": {"思想政治教育"},
	}
	clubNames = []string{"青年志愿者协会", "计算机协会", "辩论社", "摄影社", "羽毛球社", "话剧社", "创业联盟", "汉服社"}
	teamNames = []string{
		"东苑自治委员会", "西苑自治委员会", "南苑自治委员会",
		"星火志愿服务队", "蒲公英志愿服务队", "微光志愿服务队",
		"宿舍值班一队", "宿舍值班二队", "图书馆值班队", "晚自习督查队",
	}
	jobTitles = []string{
		"图书馆图书整理员", "教学楼卫生维护", "实验室助理", "校史馆讲解员",
		"招生办接待员", "宿舍管理助理", "校园广播站编辑", "新媒体内容运营",
		"网络中心运维助理", "食堂秩序维护员", "体育馆器材管理", "档案馆档案整理",
		"校园活动摄影师", "国际交流处翻译", "学工部数据助理", "心理咨询中心助理",
		"创新创业基地助理", "勤工助学服务中心助理", "校医院前台", "保卫处巡逻协助",
	}
	jobDepartments = []string{"图书馆", "后勤处", "学工部", "教务处", "宣传部", "信息中心", "保卫处", "校医院", "招生就业处"}
	locations      = []string{"图书馆一楼", "图书馆三楼", "1号教学楼", "2号教学楼", "信息中心机房", "体育馆", "学生活动中心", "行政楼"}
	categories     = []string{"社区服务", "环保公益", "敬老助残", "教育帮扶", "赛会服务", "文化宣传"}
	contactPhones  = []string{"13800001111", "13911112222", "13722223333", "15800004444", "18600005555"}

	stages = []string{
		models.StageApplicant,
		models.StageActivist,
		models.StageDevelopTarget,
		models.StagePoliticalReview,
		models.StageLeagueMember,
	}
)

// ---- 工具函数 ----

func pick[T any](rng *rand.Rand, list []T) T {
	return list[rng.Intn(len(list))]
}

func randDateStr(rng *rand.Rand, daysBack int) string {
	d := time.Now().AddDate(0, 0, -rng.Intn(daysBack))
	return d.Format("2006-01-02")
}

func randName(rng *rand.Rand) string {
	if rng.Intn(2) == 0 {
		return pick(rng, surnames) + pick(rng, names)
	}
	return pick(rng, surnames) + pick(rng, names) + pick(rng, names)
}

func ptrFloat64(v float64) *float64 { return &v }
func ptrUint(v uint) *uint           { return &v }
func ptrTime(t time.Time) *time.Time { return &t }

// ---- 主流程 ----

func main() {
	var (
		dbPath = flag.String("db", "data/unistep.db", "SQLite 数据库文件路径")
		reset  = flag.Bool("reset", false, "执行前清空业务数据表（保留 admin 等基础账号）")
	)
	flag.Parse()

	db, err := database.Connect(*dbPath)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	if err := database.Seed(db); err != nil {
		log.Fatalf("初始化基础数据失败: %v", err)
	}

	if *reset {
		if err := resetBusinessTables(db); err != nil {
			log.Fatalf("清空业务表失败: %v", err)
		}
		log.Println("已清空业务表")
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	users, err := seedUsers(db, rng, 45) // 加上 5 个内置账号约 50 人
	if err != nil {
		log.Fatalf("生成用户失败: %v", err)
	}
	log.Printf("用户总数：%d", len(users))

	// 分拣角色
	var students []models.User
	var cadres []models.User
	var teachers []models.User
	var admins []models.User
	for _, u := range users {
		switch {
		case containsRole(u.Roles, "admin"):
			admins = append(admins, u)
		case containsRole(u.Roles, "teacher"):
			teachers = append(teachers, u)
		case containsRole(u.Roles, "student_cadre"):
			cadres = append(cadres, u)
		default:
			students = append(students, u)
		}
	}
	// 学生干部也可以参与学生事务
	studentPool := append(append([]models.User{}, students...), cadres...)

	if err := seedMembers(db, rng, studentPool); err != nil {
		log.Fatalf("生成团员档案失败: %v", err)
	}
	log.Println("已生成团员档案")

	if err := seedActivities(db, rng, append(teachers, cadres...), studentPool); err != nil {
		log.Fatalf("生成社团活动失败: %v", err)
	}
	log.Println("已生成社团活动")

	if err := seedCommunity(db, rng, append(teachers, cadres...), studentPool); err != nil {
		log.Fatalf("生成社区队伍失败: %v", err)
	}
	log.Println("已生成社区队伍/值班/志愿服务")

	if err := seedWorkStudy(db, rng, append(teachers, admins...), studentPool); err != nil {
		log.Fatalf("生成勤工助学失败: %v", err)
	}
	log.Println("已生成勤工助学岗位")

	log.Println("测试数据生成完成 ✅")
}

// containsRole 简易判断 roles 字符串是否包含某个角色
func containsRole(roles, target string) bool {
	for i := 0; i < len(roles); {
		j := i
		for j < len(roles) && roles[j] != ',' {
			j++
		}
		if roles[i:j] == target {
			return true
		}
		i = j + 1
	}
	return false
}

// ---- reset ----

func resetBusinessTables(db *gorm.DB) error {
	tables := []string{
		"work_study_files", "salary_records", "work_attendances", "job_applications", "work_study_jobs",
		"volunteer_services", "duty_records", "duty_schedules", "team_members", "community_teams",
		"activity_files", "activity_checkins", "activity_registrations", "club_activities",
		"member_attachments", "political_reviews", "develop_target_records", "activist_records",
		"league_applications", "member_profiles",
	}
	for _, t := range tables {
		if err := db.Exec(fmt.Sprintf("DELETE FROM %s", t)).Error; err != nil {
			return err
		}
	}
	// 清除非内置用户（保留 admin/teacher_wang/student_li/zhang/liu）
	return db.Exec(`DELETE FROM users WHERE username NOT IN ('admin','teacher_wang','student_li','student_zhang','student_liu')`).Error
}

// ---- 用户 ----

func seedUsers(db *gorm.DB, rng *rand.Rand, n int) ([]models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte("test123"), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	hashed := string(hash)

	for i := 1; i <= n; i++ {
		username := fmt.Sprintf("stu_%04d", i)
		var exist int64
		db.Model(&models.User{}).Where("username = ?", username).Count(&exist)
		if exist > 0 {
			continue
		}

		college := pick(rng, colleges)
		role := "student"
		if i%10 == 0 {
			role = "student_cadre"
		}
		if i%15 == 0 {
			role = "teacher"
		}

		className := ""
		if role != "teacher" {
			grade := 2021 + rng.Intn(4)
			className = fmt.Sprintf("%s%02d%02d", abbreviate(college), grade%100, 1+rng.Intn(3))
		}

		u := models.User{
			Username:  username,
			Password:  hashed,
			Email:     fmt.Sprintf("%s@unistep.edu.cn", username),
			Roles:     role,
			RealName:  randName(rng),
			College:   college,
			ClassName: className,
			Status:    "active",
		}
		if err := db.Create(&u).Error; err != nil {
			return nil, err
		}
	}

	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// 学院缩写，用于班级编号
func abbreviate(college string) string {
	switch college {
	case "计算机学院":
		return "计科"
	case "电子信息学院":
		return "电子"
	case "机械工程学院":
		return "机械"
	case "经济管理学院":
		return "经管"
	case "外国语学院":
		return "外语"
	case "马克思主义学院":
		return "马院"
	}
	return "通用"
}

// ---- 团员档案 ----

func seedMembers(db *gorm.DB, rng *rand.Rand, students []models.User) error {
	if len(students) == 0 {
		return nil
	}
	count := 50
	if len(students) < count {
		count = len(students)
	}

	for i := 0; i < count; i++ {
		u := students[i]
		var exist int64
		db.Model(&models.MemberProfile{}).Where("user_id = ?", u.ID).Count(&exist)
		if exist > 0 {
			continue
		}

		stage := stages[i%len(stages)]
		studentNo := fmt.Sprintf("2023%07d", 1000+i)
		majorList := majors[u.College]
		if len(majorList) == 0 {
			majorList = []string{"通用专业"}
		}
		gender := "男"
		if rng.Intn(2) == 0 {
			gender = "女"
		}

		profile := models.MemberProfile{
			UserID:    u.ID,
			Name:      u.RealName,
			StudentNo: studentNo,
			Gender:    gender,
			Birthday:  fmt.Sprintf("200%d-%02d-%02d", 2+rng.Intn(4), 1+rng.Intn(12), 1+rng.Intn(28)),
			IDCard:    fmt.Sprintf("3201%011d", 10000000000+rng.Int63n(89999999999)),
			Nation:    "汉",
			Phone:     fmt.Sprintf("13%09d", 100000000+rng.Intn(899999999)),
			College:   u.College,
			Major:     pick(rng, majorList),
			ClassName: u.ClassName,
			Stage:     stage,
			JoinDate:  randDateStr(rng, 730),
			Remark:    "测试数据自动生成",
		}
		if err := db.Create(&profile).Error; err != nil {
			return err
		}

		// 入团申请（所有阶段都有）
		application := models.LeagueApplication{
			ProfileID:  profile.ID,
			ApplyDate:  randDateStr(rng, 720),
			Motivation: "拥护中国共产党，积极向党组织靠拢，希望加入中国共产主义青年团。",
			Introducer: randName(rng),
			Status:     "approved",
			ReviewNote: "材料完整，符合要求",
		}
		if stage == models.StageApplicant {
			application.Status = "pending"
			application.ReviewNote = ""
		}
		if err := db.Create(&application).Error; err != nil {
			return err
		}

		// 积极分子记录
		if stage == models.StageActivist || stage == models.StageDevelopTarget ||
			stage == models.StagePoliticalReview || stage == models.StageLeagueMember {
			ar := models.ActivistRecord{
				ProfileID:  profile.ID,
				StartDate:  randDateStr(rng, 540),
				Trainer:    randName(rng) + "老师",
				TrainPlan:  "参加青年大学习、团课培训、志愿服务，每月汇报一次思想动态。",
				Evaluation: "学习积极，表现良好。",
				Score:      80 + rng.Float32()*15,
			}
			if err := db.Create(&ar).Error; err != nil {
				return err
			}
		}

		// 发展对象记录
		if stage == models.StageDevelopTarget || stage == models.StagePoliticalReview || stage == models.StageLeagueMember {
			dt := models.DevelopTargetRecord{
				ProfileID:     profile.ID,
				ConfirmedDate: randDateStr(rng, 365),
				Mentor:        randName(rng) + "老师",
				PublicityNote: "已在班级及学院公示 7 天，无异议。",
				Conclusion:    "通过",
			}
			if err := db.Create(&dt).Error; err != nil {
				return err
			}
		}

		// 政审记录
		if stage == models.StagePoliticalReview || stage == models.StageLeagueMember {
			pr := models.PoliticalReview{
				ProfileID:     profile.ID,
				ReviewDate:    randDateStr(rng, 180),
				Reviewer:      randName(rng) + "老师",
				FamilyMembers: "父亲：某某，群众；母亲：某某，群众。",
				Conclusion:    "审核通过",
				Status:        "filed",
			}
			if err := db.Create(&pr).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// ---- 社团活动 ----

func seedActivities(db *gorm.DB, rng *rand.Rand, organizers, students []models.User) error {
	if len(organizers) == 0 || len(students) == 0 {
		return nil
	}
	activityStatuses := []string{
		models.ActivityDraft, models.ActivityPending, models.ActivityRejected,
		models.ActivityRegOpen, models.ActivityRegClosed,
		models.ActivityInProgress, models.ActivityCompleted, models.ActivityArchived,
	}

	for i := 0; i < 20; i++ {
		status := activityStatuses[i%len(activityStatuses)]
		start := time.Now().AddDate(0, 0, rng.Intn(60)-30)
		end := start.Add(time.Duration(2+rng.Intn(6)) * time.Hour)

		club := pick(rng, clubNames)
		title := fmt.Sprintf("%s%s%d期", club, []string{"主题沙龙", "技能培训", "户外实践", "公益活动", "纳新宣讲"}[rng.Intn(5)], i+1)

		budget := ptrFloat64(float64(500 + rng.Intn(4500)))
		var approvedBy *uint
		var approvedAt *time.Time
		if status != models.ActivityDraft && status != models.ActivityPending {
			approver := pick(rng, organizers)
			approvedBy = ptrUint(approver.ID)
			approvedAt = ptrTime(start.AddDate(0, 0, -7))
		}

		act := models.ClubActivity{
			ClubName:    club,
			Title:       title,
			StartTime:   start,
			EndTime:     end,
			Location:    pick(rng, locations),
			Capacity:    20 + rng.Intn(80),
			Description: "本次活动旨在丰富同学课余生活，提升综合素质，欢迎积极报名参加。",
			Budget:      budget,
			Status:      status,
			Summary:     "",
			CreatedBy:   pick(rng, organizers).ID,
			ApprovedBy:  approvedBy,
			ApprovedAt:  approvedAt,
		}
		if status == models.ActivityCompleted || status == models.ActivityArchived {
			act.Summary = "活动圆满完成，达到预期效果，参与同学反响热烈。"
		}
		if status == models.ActivityRejected {
			act.ApprovalOpinion = "活动方案需进一步完善，请补充安全预案。"
		}
		if err := db.Create(&act).Error; err != nil {
			return err
		}

		// 仅对开放报名以后的活动生成报名记录
		if status == models.ActivityRegOpen || status == models.ActivityRegClosed ||
			status == models.ActivityInProgress || status == models.ActivityCompleted ||
			status == models.ActivityArchived {

			regCount := 5 + rng.Intn(15)
			if regCount > len(students) {
				regCount = len(students)
			}
			// 打乱学生顺序避免重复
			perm := rng.Perm(len(students))
			for j := 0; j < regCount; j++ {
				stu := students[perm[j]]
				reg := models.ActivityRegistration{
					ActivityID:   act.ID,
					StudentID:    stu.ID,
					Status:       "registered",
					RegisteredAt: start.AddDate(0, 0, -3),
				}
				if rng.Intn(10) == 0 {
					reg.Status = "cancelled"
					reg.CancelledAt = ptrTime(start.AddDate(0, 0, -1))
				}
				if err := db.Create(&reg).Error; err != nil {
					return err
				}

				// 签到只在进行中或已完成的活动生成
				if (status == models.ActivityInProgress || status == models.ActivityCompleted || status == models.ActivityArchived) &&
					reg.Status == "registered" && rng.Intn(10) < 8 {
					method := "qr"
					if rng.Intn(3) == 0 {
						method = "manual"
					}
					checkin := models.ActivityCheckin{
						ActivityID:    act.ID,
						StudentID:     stu.ID,
						CheckinTime:   start.Add(time.Duration(rng.Intn(30)) * time.Minute),
						CheckinMethod: method,
					}
					if err := db.Create(&checkin).Error; err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// ---- 社区队伍 / 值班 / 志愿服务 ----

func seedCommunity(db *gorm.DB, rng *rand.Rand, leaders, students []models.User) error {
	if len(leaders) == 0 || len(students) == 0 {
		return nil
	}
	teamTypes := []string{models.TeamTypeAutonomy, models.TeamTypeVolunteer, models.TeamTypeDuty}

	for i, name := range teamNames {
		teamType := teamTypes[i%len(teamTypes)]
		creator := pick(rng, leaders)
		team := models.CommunityTeam{
			Name:        name,
			TeamType:    teamType,
			Description: name + "由学生骨干自主组织开展日常事务管理与服务。",
			Quota:       10 + rng.Intn(15),
			Location:    pick(rng, locations),
			ContactInfo: pick(rng, contactPhones),
			Status:      "active",
			CreatedBy:   creator.ID,
		}
		if err := db.Create(&team).Error; err != nil {
			return err
		}

		// 成员
		memberCount := 6 + rng.Intn(8)
		perm := rng.Perm(len(students))
		var teamMembers []models.TeamMember
		for j := 0; j < memberCount && j < len(students); j++ {
			stu := students[perm[j]]
			role := models.MemberRoleMember
			if j == 0 {
				role = models.MemberRoleLeader
			} else if j == 1 {
				role = models.MemberRoleVice
			}
			tm := models.TeamMember{
				TeamID:    team.ID,
				UserID:    stu.ID,
				Name:      stu.RealName,
				StudentNo: fmt.Sprintf("2023%07d", 2000+stu.ID),
				Role:      role,
				Status:    models.MemberStatusActive,
				JoinDate:  randDateStr(rng, 365),
				TermStart: "2024-09",
				TermEnd:   "2025-06",
				Remark:    "",
			}
			if err := db.Create(&tm).Error; err != nil {
				return err
			}
			teamMembers = append(teamMembers, tm)
		}

		// 值班排班（仅 duty 队伍）
		if teamType == models.TeamTypeDuty {
			for d := 0; d < 6; d++ {
				date := time.Now().AddDate(0, 0, d-3)
				schedule := models.DutySchedule{
					TeamID:    team.ID,
					Date:      date.Format("2006-01-02"),
					StartTime: "19:00",
					EndTime:   "22:00",
					Location:  team.Location,
					Status:    models.DutyStatusScheduled,
					CreatedBy: creator.ID,
				}
				// 历史班次设为完成
				if d < 3 {
					schedule.Status = models.DutyStatusCompleted
				}
				if err := db.Create(&schedule).Error; err != nil {
					return err
				}

				// 排班生成 2~3 条值班记录
				recCount := 2 + rng.Intn(2)
				rperm := rng.Perm(len(teamMembers))
				for k := 0; k < recCount && k < len(teamMembers); k++ {
					m := teamMembers[rperm[k]]
					checkin := time.Date(date.Year(), date.Month(), date.Day(), 19, rng.Intn(15), 0, 0, time.Local)
					rec := models.DutyRecord{
						ScheduleID:  schedule.ID,
						TeamID:      team.ID,
						UserID:      m.UserID,
						Name:        m.Name,
						CheckinTime: ptrTime(checkin),
						Status:      models.DutyStatusScheduled,
						Remark:      "",
					}
					if schedule.Status == models.DutyStatusCompleted {
						checkout := checkin.Add(time.Duration(150+rng.Intn(60)) * time.Minute)
						duration := checkout.Sub(checkin).Hours()
						rec.CheckoutTime = ptrTime(checkout)
						rec.Duration = ptrFloat64(duration)
						rec.Status = models.DutyStatusCompleted
					}
					if err := db.Create(&rec).Error; err != nil {
						return err
					}
				}
			}
		}

		// 志愿服务（仅 volunteer 队伍）
		if teamType == models.TeamTypeVolunteer {
			for s := 0; s < 8; s++ {
				m := teamMembers[rng.Intn(len(teamMembers))]
				hours := 1.5 + rng.Float64()*3
				verified := rng.Intn(10) < 8
				vs := models.VolunteerService{
					TeamID:      team.ID,
					UserID:      m.UserID,
					Name:        m.Name,
					StudentNo:   m.StudentNo,
					Title:       team.Name + "第" + fmt.Sprintf("%d", s+1) + "次服务",
					Date:        randDateStr(rng, 90),
					Hours:       hours,
					Category:    pick(rng, categories),
					Description: "在指定地点提供志愿服务，包括秩序维护、咨询解答、清洁等。",
					Verified:    verified,
					CreatedBy:   creator.ID,
				}
				if verified {
					vs.VerifiedBy = ptrUint(creator.ID)
					vs.VerifiedAt = ptrTime(time.Now().AddDate(0, 0, -rng.Intn(30)))
				}
				if err := db.Create(&vs).Error; err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// ---- 勤工助学 ----

func seedWorkStudy(db *gorm.DB, rng *rand.Rand, admins, students []models.User) error {
	if len(admins) == 0 || len(students) == 0 {
		return nil
	}
	jobStatuses := []string{models.JobDraft, models.JobPublished, models.JobClosed, models.JobCompleted}
	appStatuses := []string{models.AppApplied, models.AppAccepted, models.AppRejected, models.AppCancelled}

	for i := 0; i < 20; i++ {
		status := jobStatuses[i%len(jobStatuses)]
		start := time.Now().AddDate(0, 0, -rng.Intn(60))
		end := start.AddDate(0, rng.Intn(3)+1, 0)
		creator := pick(rng, admins)

		job := models.WorkStudyJob{
			Title:         jobTitles[i%len(jobTitles)],
			Department:    pick(rng, jobDepartments),
			Location:      pick(rng, locations),
			Description:   "本岗位要求工作认真负责，能够按时完成排班任务，鼓励学生在实践中锻炼能力。",
			Quota:         2 + rng.Intn(8),
			SalaryPerHour: 15.0 + float64(rng.Intn(15)),
			StartTime:     start,
			EndTime:       end,
			ContactPerson: randName(rng) + "老师",
			ContactPhone:  pick(rng, contactPhones),
			Status:        status,
			CreatedBy:     creator.ID,
		}
		if err := db.Create(&job).Error; err != nil {
			return err
		}

		// 草稿岗位不生成报名
		if status == models.JobDraft {
			continue
		}

		appCount := 4 + rng.Intn(8)
		if appCount > len(students) {
			appCount = len(students)
		}
		perm := rng.Perm(len(students))
		for j := 0; j < appCount; j++ {
			stu := students[perm[j]]
			appStatus := appStatuses[rng.Intn(len(appStatuses))]
			appliedAt := start.AddDate(0, 0, -3-rng.Intn(7))
			app := models.JobApplication{
				JobID:     job.ID,
				StudentID: stu.ID,
				Status:    appStatus,
				Remark:    "",
				AppliedAt: appliedAt,
			}
			switch appStatus {
			case models.AppAccepted:
				app.AcceptedAt = ptrTime(appliedAt.AddDate(0, 0, 2))
			case models.AppRejected:
				app.RejectedAt = ptrTime(appliedAt.AddDate(0, 0, 2))
				app.Remark = "岗位已满，欢迎下次再投。"
			case models.AppCancelled:
				app.CancelledAt = ptrTime(appliedAt.AddDate(0, 0, 1))
			}
			if err := db.Create(&app).Error; err != nil {
				return err
			}

			// 已录用的学生生成考勤与薪资
			if appStatus == models.AppAccepted && (status == models.JobPublished || status == models.JobClosed || status == models.JobCompleted) {
				totalHours := 0.0
				for d := 0; d < 4+rng.Intn(6); d++ {
					date := time.Now().AddDate(0, 0, -d*2)
					checkin := time.Date(date.Year(), date.Month(), date.Day(), 9+rng.Intn(2), rng.Intn(30), 0, 0, time.Local)
					checkout := checkin.Add(time.Duration(2+rng.Intn(3)) * time.Hour)
					hours := checkout.Sub(checkin).Hours()
					totalHours += hours
					att := models.WorkAttendance{
						JobID:        job.ID,
						StudentID:    stu.ID,
						Date:         date.Format("2006-01-02"),
						CheckinTime:  checkin,
						CheckoutTime: ptrTime(checkout),
						Hours:        hours,
						Method:       "qr",
					}
					if rng.Intn(4) == 0 {
						att.Method = "manual"
					}
					if err := db.Create(&att).Error; err != nil {
						return err
					}
				}

				// 月度薪资记录
				salaryStatus := models.SalaryPending
				var paidAt *time.Time
				if status == models.JobCompleted || rng.Intn(2) == 0 {
					salaryStatus = models.SalaryPaid
					paidAt = ptrTime(time.Now().AddDate(0, 0, -rng.Intn(20)))
				}
				sal := models.SalaryRecord{
					JobID:     job.ID,
					StudentID: stu.ID,
					Month:     time.Now().Format("2006-01"),
					Hours:     totalHours,
					Amount:    totalHours * job.SalaryPerHour,
					Status:    salaryStatus,
					PaidAt:    paidAt,
				}
				if err := db.Create(&sal).Error; err != nil {
					return err
				}
			}
		}
	}
	return nil
}
