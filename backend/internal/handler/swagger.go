package handler

import (
	"github.com/gin-gonic/gin"
)

// swaggerSpec 是手写的 OpenAPI 3.0 文档；与 docs/api-spec.md 保持一致
var swaggerSpec = map[string]any{
	"openapi": "3.0.3",
	"info": map[string]any{
		"title":       "UniStep Platform API",
		"version":     "1.0.0",
		"description": "学生一站式自主管理过程管理系统接口文档（Sprint6：勤工助学模块）",
	},
	"servers": []map[string]any{
		{"url": "/", "description": "当前服务"},
	},
	"components": map[string]any{
		"securitySchemes": map[string]any{
			"bearerAuth": map[string]any{
				"type":         "http",
				"scheme":       "bearer",
				"bearerFormat": "JWT",
			},
		},
		"schemas": map[string]any{
			"ApiResponse": objSchema(map[string]any{
				"code":    strSchema(),
				"message": strSchema(),
				"data":    map[string]any{"type": "object"},
			}),
			"MemberProfileRequest": objSchema(map[string]any{
				"userId":    intSchema(),
				"name":      strSchema(),
				"studentNo": strSchema(),
				"gender":    strSchema(),
				"birthday":  strSchema(),
				"idCard":    strSchema(),
				"nation":    strSchema(),
				"phone":     strSchema(),
				"college":   strSchema(),
				"major":     strSchema(),
				"className": strSchema(),
				"stage":     strSchema(),
				"joinDate":  strSchema(),
				"remark":    strSchema(),
			}),
			"LeagueApplication": objSchema(map[string]any{
				"applyDate":  strSchema(),
				"motivation": strSchema(),
				"introducer": strSchema(),
				"status":     strSchema(),
				"reviewNote": strSchema(),
			}),
			"ActivistRecord": objSchema(map[string]any{
				"startDate":  strSchema(),
				"trainer":    strSchema(),
				"trainPlan":  strSchema(),
				"evaluation": strSchema(),
				"score":      map[string]any{"type": "number"},
			}),
			"DevelopTargetRecord": objSchema(map[string]any{
				"confirmedDate": strSchema(),
				"mentor":        strSchema(),
				"publicityNote": strSchema(),
				"conclusion":    strSchema(),
			}),
			"PoliticalReview": objSchema(map[string]any{
				"reviewDate":    strSchema(),
				"reviewer":      strSchema(),
				"familyMembers": strSchema(),
				"conclusion":    strSchema(),
				"status":        strSchema(),
			}),
			"ActivityRequest": objSchema(map[string]any{
				"clubName":    strSchema(),
				"title":       strSchema(),
				"startTime":   strSchema(),
				"endTime":     strSchema(),
				"location":    strSchema(),
				"capacity":    intSchema(),
				"description": strSchema(),
				"budget":      map[string]any{"type": "number"},
			}),
			"ApprovalRequest": objSchema(map[string]any{
				"opinion": strSchema(),
				"approve": map[string]any{"type": "boolean"},
			}),
			"CheckinRequest": objSchema(map[string]any{
				"studentId": intSchema(),
			}),
			"SummaryRequest": objSchema(map[string]any{
				"summary": strSchema(),
			}),
			"StatusUpdateRequest": objSchema(map[string]any{
				"status": strSchema(),
			}),
			"TeamRequest": objSchema(map[string]any{
				"name":        strSchema(),
				"teamType":    strSchema(),
				"description": strSchema(),
				"quota":       intSchema(),
				"location":    strSchema(),
				"contactInfo": strSchema(),
			}),
			"TeamMemberRequest": objSchema(map[string]any{
				"userId":    intSchema(),
				"name":      strSchema(),
				"studentNo": strSchema(),
				"role":      strSchema(),
				"joinDate":  strSchema(),
				"termStart": strSchema(),
				"termEnd":   strSchema(),
				"remark":    strSchema(),
			}),
			"DutyScheduleRequest": objSchema(map[string]any{
				"date":      strSchema(),
				"startTime": strSchema(),
				"endTime":   strSchema(),
				"location":  strSchema(),
				"memberIds": map[string]any{"type": "array", "items": intSchema()},
			}),
			"DutyCheckinRequest": objSchema(map[string]any{
				"userId": intSchema(),
			}),
			"DutyCheckoutRequest": objSchema(map[string]any{
				"userId": intSchema(),
			}),
			"VolunteerServiceRequest": objSchema(map[string]any{
				"userId":      intSchema(),
				"name":        strSchema(),
				"studentNo":   strSchema(),
				"title":       strSchema(),
				"date":        strSchema(),
				"hours":       map[string]any{"type": "number"},
				"category":    strSchema(),
				"description": strSchema(),
			}),
			"VerifyServiceRequest": objSchema(map[string]any{
				"verified": map[string]any{"type": "boolean"},
			}),
			"JobRequest": objSchema(map[string]any{
				"title":         strSchema(),
				"department":    strSchema(),
				"location":      strSchema(),
				"description":   strSchema(),
				"quota":         intSchema(),
				"salaryPerHour": map[string]any{"type": "number"},
				"startTime":     strSchema(),
				"endTime":       strSchema(),
				"contactPerson": strSchema(),
				"contactPhone":  strSchema(),
			}),
			"AttendanceRequest": objSchema(map[string]any{
				"studentId": intSchema(),
				"date":      strSchema(),
				"method":    strSchema(),
			}),
			"SalaryCalculateRequest": objSchema(map[string]any{
				"month": strSchema(),
			}),
			"ApplicationActionRequest": objSchema(map[string]any{
				"remark": strSchema(),
			}),
		},
	},
	"security": []map[string]any{
		{"bearerAuth": []string{}},
	},
	"paths": map[string]any{
		"/api/v1/auth/register": map[string]any{
			"post": op("用户注册", "auth", nil, "RegisterRequest"),
		},
		"/api/v1/auth/login": map[string]any{
			"post": op("用户登录", "auth", nil, "LoginRequest"),
		},
		"/api/v1/members": map[string]any{
			"get":  op("团员列表", "团员发展", []string{"page", "size", "stage", "name"}, ""),
			"post": op("创建团员档案", "团员发展", nil, "MemberProfileRequest"),
		},
		"/api/v1/members/{id}": map[string]any{
			"get":    op("获取团员档案", "团员发展", []string{"id"}, ""),
			"put":    op("更新团员档案", "团员发展", []string{"id"}, "MemberProfileRequest"),
			"delete": op("删除团员档案", "团员发展", []string{"id"}, ""),
		},
		"/api/v1/members/{id}/applications": map[string]any{
			"post": op("提交入团申请", "团员发展", []string{"id"}, "LeagueApplication"),
		},
		"/api/v1/members/{id}/activists": map[string]any{
			"post": op("录入积极分子培养记录", "团员发展", []string{"id"}, "ActivistRecord"),
		},
		"/api/v1/members/{id}/develop-targets": map[string]any{
			"post": op("录入发展对象记录", "团员发展", []string{"id"}, "DevelopTargetRecord"),
		},
		"/api/v1/members/{id}/political-reviews": map[string]any{
			"post": op("政审备案", "团员发展", []string{"id"}, "PoliticalReview"),
		},
		"/api/v1/members/{id}/attachments": map[string]any{
			"post": map[string]any{
				"summary": "上传档案附件（multipart/form-data）",
				"tags":    []string{"团员发展"},
				"parameters": []map[string]any{
					{"name": "id", "in": "path", "required": true, "schema": intSchema()},
				},
				"requestBody": map[string]any{
					"required": true,
					"content": map[string]any{
						"multipart/form-data": map[string]any{
							"schema": objSchema(map[string]any{
								"category": strSchema(),
								"file":     map[string]any{"type": "string", "format": "binary"},
							}),
						},
					},
				},
				"responses": map[string]any{
					"201": map[string]any{
						"description": "上传成功",
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{"$ref": "#/components/schemas/ApiResponse"},
							},
						},
					},
				},
			},
		},
		"/api/v1/members/{id}/archive": map[string]any{
			"get": op("生成团员电子档案", "团员发展", []string{"id"}, ""),
		},

		// ---- 社团活动模块 ----
		"/api/v1/activities": map[string]any{
			"get":  op("活动列表", "社团活动", []string{"page", "size", "status", "clubName", "title"}, ""),
			"post": op("创建活动", "社团活动", nil, "ActivityRequest"),
		},
		"/api/v1/activities/statistics": map[string]any{
			"get": op("活动统计", "社团活动", nil, ""),
		},
		"/api/v1/activities/{id}": map[string]any{
			"get":    op("获取活动详情", "社团活动", []string{"id"}, ""),
			"put":    op("更新活动", "社团活动", []string{"id"}, "ActivityRequest"),
			"delete": op("删除活动", "社团活动", []string{"id"}, ""),
		},
		"/api/v1/activities/{id}/submit": map[string]any{
			"post": op("提交审批", "社团活动", []string{"id"}, ""),
		},
		"/api/v1/activities/{id}/approve": map[string]any{
			"post": op("审批活动", "社团活动", []string{"id"}, "ApprovalRequest"),
		},
		"/api/v1/activities/{id}/register": map[string]any{
			"post": op("活动报名", "社团活动", []string{"id"}, ""),
		},
		"/api/v1/activities/{id}/cancel-registration": map[string]any{
			"post": op("取消报名", "社团活动", []string{"id"}, ""),
		},
		"/api/v1/activities/{id}/checkin": map[string]any{
			"post": op("活动签到", "社团活动", []string{"id"}, "CheckinRequest"),
		},
		"/api/v1/activities/{id}/files": map[string]any{
			"post": map[string]any{
				"summary": "上传活动图片/文件（multipart/form-data）",
				"tags":    []string{"社团活动"},
				"parameters": []map[string]any{
					{"name": "id", "in": "path", "required": true, "schema": intSchema()},
				},
				"requestBody": map[string]any{
					"required": true,
					"content": map[string]any{
						"multipart/form-data": map[string]any{
							"schema": objSchema(map[string]any{
								"fileType": strSchema(),
								"file":     map[string]any{"type": "string", "format": "binary"},
							}),
						},
					},
				},
				"responses": map[string]any{
					"201": map[string]any{
						"description": "上传成功",
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{"$ref": "#/components/schemas/ApiResponse"},
							},
						},
					},
				},
			},
		},
		"/api/v1/activities/{id}/summary": map[string]any{
			"post": op("提交活动总结", "社团活动", []string{"id"}, "SummaryRequest"),
		},
		"/api/v1/activities/{id}/status": map[string]any{
			"put": op("变更活动状态", "社团活动", []string{"id"}, "StatusUpdateRequest"),
		},

		// ---- 学生社区与自治队伍模块 ----
		"/api/v1/community/teams": map[string]any{
			"get":  op("队伍列表", "社区队伍", []string{"page", "size", "teamType", "name", "status"}, ""),
			"post": op("创建队伍", "社区队伍", nil, "TeamRequest"),
		},
		"/api/v1/community/teams/statistics": map[string]any{
			"get": op("队伍统计", "社区队伍", nil, ""),
		},
		"/api/v1/community/service-profile": map[string]any{
			"get": op("服务时长个人档案", "社区队伍", []string{"userId"}, ""),
		},
		"/api/v1/community/teams/{id}": map[string]any{
			"get":    op("获取队伍详情", "社区队伍", []string{"id"}, ""),
			"put":    op("更新队伍", "社区队伍", []string{"id"}, "TeamRequest"),
			"delete": op("解散队伍", "社区队伍", []string{"id"}, ""),
		},
		"/api/v1/community/teams/{id}/members": map[string]any{
			"get":  op("队伍成员列表", "社区队伍", []string{"id", "status", "role"}, ""),
			"post": op("添加成员（纳新）", "社区队伍", []string{"id"}, "TeamMemberRequest"),
		},
		"/api/v1/community/teams/{id}/members/{memberId}": map[string]any{
			"put":    op("更新成员信息（换届）", "社区队伍", []string{"id", "memberId"}, "TeamMemberRequest"),
			"delete": op("移除成员", "社区队伍", []string{"id", "memberId"}, ""),
		},
		"/api/v1/community/teams/{id}/duties": map[string]any{
			"get":  op("值班安排列表", "社区队伍", []string{"id", "page", "size", "date", "status"}, ""),
			"post": op("创建值班安排", "社区队伍", []string{"id"}, "DutyScheduleRequest"),
		},
		"/api/v1/community/teams/{id}/duties/{scheduleId}/checkin": map[string]any{
			"post": op("值班签到", "社区队伍", []string{"id", "scheduleId"}, "DutyCheckinRequest"),
		},
		"/api/v1/community/teams/{id}/duties/{scheduleId}/checkout": map[string]any{
			"post": op("值班签退", "社区队伍", []string{"id", "scheduleId"}, "DutyCheckoutRequest"),
		},
		"/api/v1/community/teams/{id}/services": map[string]any{
			"get":  op("志愿服务列表", "社区队伍", []string{"id", "page", "size", "category", "verified"}, ""),
			"post": op("记录志愿服务", "社区队伍", []string{"id"}, "VolunteerServiceRequest"),
		},
		"/api/v1/community/teams/{id}/services/{serviceId}/verify": map[string]any{
			"put": op("核实志愿服务", "社区队伍", []string{"id", "serviceId"}, "VerifyServiceRequest"),
		},

		// ---- 勤工助学模块 ----
		"/api/v1/workstudy/statistics": map[string]any{
			"get": op("勤工助学统计", "勤工助学", nil, ""),
		},
		"/api/v1/workstudy/jobs": map[string]any{
			"get":  op("岗位列表", "勤工助学", []string{"page", "size", "status", "department", "title"}, ""),
			"post": op("创建岗位", "勤工助学", nil, "JobRequest"),
		},
		"/api/v1/workstudy/jobs/{id}": map[string]any{
			"get":    op("获取岗位详情", "勤工助学", []string{"id"}, ""),
			"put":    op("更新岗位", "勤工助学", []string{"id"}, "JobRequest"),
			"delete": op("删除岗位", "勤工助学", []string{"id"}, ""),
		},
		"/api/v1/workstudy/jobs/{id}/publish": map[string]any{
			"post": op("发布岗位", "勤工助学", []string{"id"}, ""),
		},
		"/api/v1/workstudy/jobs/{id}/close": map[string]any{
			"post": op("关闭岗位", "勤工助学", []string{"id"}, ""),
		},
		"/api/v1/workstudy/jobs/{id}/apply": map[string]any{
			"post": op("学生报名", "勤工助学", []string{"id"}, ""),
		},
		"/api/v1/workstudy/jobs/{id}/cancel-application": map[string]any{
			"post": op("取消报名", "勤工助学", []string{"id"}, ""),
		},
		"/api/v1/workstudy/jobs/{id}/applications": map[string]any{
			"get": op("报名列表", "勤工助学", []string{"id", "status"}, ""),
		},
		"/api/v1/workstudy/jobs/{id}/applications/{appId}/accept": map[string]any{
			"post": op("录用学生", "勤工助学", []string{"id", "appId"}, "ApplicationActionRequest"),
		},
		"/api/v1/workstudy/jobs/{id}/applications/{appId}/reject": map[string]any{
			"post": op("拒绝报名", "勤工助学", []string{"id", "appId"}, "ApplicationActionRequest"),
		},
		"/api/v1/workstudy/jobs/{id}/attendances": map[string]any{
			"get":  op("考勤列表", "勤工助学", []string{"id", "studentId", "date"}, ""),
			"post": op("创建考勤", "勤工助学", []string{"id"}, "AttendanceRequest"),
		},
		"/api/v1/workstudy/attendances/{attId}/checkout": map[string]any{
			"put": op("考勤签退", "勤工助学", []string{"attId"}, ""),
		},
		"/api/v1/workstudy/jobs/{id}/salary/calculate": map[string]any{
			"post": op("计算薪资", "勤工助学", []string{"id"}, "SalaryCalculateRequest"),
		},
		"/api/v1/workstudy/jobs/{id}/salary": map[string]any{
			"get": op("薪资列表", "勤工助学", []string{"id", "month", "status", "studentId"}, ""),
		},
		"/api/v1/workstudy/salary/{salaryId}/pay": map[string]any{
			"put": op("发放薪资", "勤工助学", []string{"salaryId"}, ""),
		},
		"/api/v1/workstudy/jobs/{id}/files": map[string]any{
			"post": map[string]any{
				"summary": "上传岗位附件（multipart/form-data）",
				"tags":    []string{"勤工助学"},
				"parameters": []map[string]any{
					{"name": "id", "in": "path", "required": true, "schema": intSchema()},
				},
				"requestBody": map[string]any{
					"required": true,
					"content": map[string]any{
						"multipart/form-data": map[string]any{
							"schema": objSchema(map[string]any{
								"fileType": strSchema(),
								"file":     map[string]any{"type": "string", "format": "binary"},
							}),
						},
					},
				},
				"responses": map[string]any{
					"201": map[string]any{
						"description": "上传成功",
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{"$ref": "#/components/schemas/ApiResponse"},
							},
						},
					},
				},
			},
		},
	},
}

func op(summary, tag string, pathParams []string, bodyRef string) map[string]any {
	params := []map[string]any{}
	for _, name := range pathParams {
		in := "path"
		// 已知的查询参数列在此
		switch name {
		case "page", "size", "stage", "name", "status", "clubName", "title", "teamType", "date", "category", "verified", "userId", "role", "memberId", "scheduleId", "serviceId", "department", "appId", "attId", "salaryId", "month":
			in = "query"
		}
		schema := strSchema()
		if name == "id" || name == "page" || name == "size" || name == "userId" || name == "memberId" || name == "scheduleId" || name == "serviceId" || name == "appId" || name == "attId" || name == "salaryId" {
			schema = intSchema()
		}
		params = append(params, map[string]any{
			"name":     name,
			"in":       in,
			"required": in == "path",
			"schema":   schema,
		})
	}

	o := map[string]any{
		"summary": summary,
		"tags":    []string{tag},
		"responses": map[string]any{
			"200": map[string]any{
				"description": "成功",
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{"$ref": "#/components/schemas/ApiResponse"},
					},
				},
			},
		},
	}
	if len(params) > 0 {
		o["parameters"] = params
	}
	if bodyRef != "" {
		o["requestBody"] = map[string]any{
			"required": true,
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{"$ref": "#/components/schemas/" + bodyRef},
				},
			},
		}
	}
	return o
}

func strSchema() map[string]any { return map[string]any{"type": "string"} }
func intSchema() map[string]any { return map[string]any{"type": "integer"} }
func objSchema(props map[string]any) map[string]any {
	return map[string]any{"type": "object", "properties": props}
}

// SwaggerJSON 暴露 OpenAPI JSON 文档
func SwaggerJSON(c *gin.Context) {
	c.JSON(200, swaggerSpec)
}

// SwaggerUI 返回基于 CDN 的 Swagger UI HTML 页面
func SwaggerUI(c *gin.Context) {
	const html = `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8" />
  <title>UniStep API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({ url: '/swagger.json', dom_id: '#swagger-ui' });
  </script>
</body>
</html>`
	c.Data(200, "text/html; charset=utf-8", []byte(html))
}
