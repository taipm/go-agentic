package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	agenticcore "github.com/taipm/go-agentic/core"
)

// QuizState tracks the state of the quiz exam
type QuizState struct {
	TotalQuestions   int              `json:"total_questions"`
	CurrentQuestion  int              `json:"current_question"`
	CorrectAnswers   int              `json:"correct_answers"`
	WrongAnswers     int              `json:"wrong_answers"`
	QuestionHistory  []QuestionRecord `json:"question_history"`
	IsComplete       bool             `json:"is_complete"`
	StudentName      string           `json:"student_name"`
	ExamTopic        string           `json:"exam_topic"`
	StartTime        time.Time        `json:"start_time"`
	ReportPath       string           `json:"report_path"`
	mu               sync.RWMutex
}

// QuestionRecord records a single question and its result
type QuestionRecord struct {
	QuestionNumber int    `json:"question_number"`
	Question       string `json:"question"`
	StudentAnswer  string `json:"student_answer"`
	IsCorrect      bool   `json:"is_correct"`
	Points         int    `json:"points"`
	TeacherComment string `json:"teacher_comment"`
}

// NewQuizState creates a new quiz state for 10 questions
func NewQuizState(outputDir string) *QuizState {
	now := time.Now()
	reportPath := filepath.Join(outputDir, fmt.Sprintf("exam_%s.md", now.Format("20060102_150405")))

	return &QuizState{
		TotalQuestions:  10,
		CurrentQuestion: 0,
		CorrectAnswers:  0,
		WrongAnswers:    0,
		QuestionHistory: make([]QuestionRecord, 0),
		IsComplete:      false,
		StudentName:     "Học sinh",
		ExamTopic:       "Kiến thức tổng hợp",
		StartTime:       now,
		ReportPath:      reportPath,
	}
}

// SetStudentInfo sets student name and exam topic
func (qs *QuizState) SetStudentInfo(name, topic string) {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	if name != "" {
		qs.StudentName = name
	}
	if topic != "" {
		qs.ExamTopic = topic
	}
}

// GenerateMarkdownReport generates the exam report as markdown
func (qs *QuizState) GenerateMarkdownReport(teacherComment string) string {
	qs.mu.RLock()
	defer qs.mu.RUnlock()

	var sb strings.Builder

	// Header
	sb.WriteString("# BIÊN BẢN THI VẤN ĐÁP\n\n")
	sb.WriteString("---\n\n")

	// Exam info
	sb.WriteString("## Thông tin kỳ thi\n\n")
	sb.WriteString(fmt.Sprintf("| Thông tin | Chi tiết |\n"))
	sb.WriteString(fmt.Sprintf("|-----------|----------|\n"))
	sb.WriteString(fmt.Sprintf("| **Thí sinh** | %s |\n", qs.StudentName))
	sb.WriteString(fmt.Sprintf("| **Chủ đề** | %s |\n", qs.ExamTopic))
	sb.WriteString(fmt.Sprintf("| **Thời gian bắt đầu** | %s |\n", qs.StartTime.Format("15:04:05 - 02/01/2006")))
	sb.WriteString(fmt.Sprintf("| **Tổng số câu hỏi** | %d |\n", qs.TotalQuestions))
	sb.WriteString("\n---\n\n")

	// Questions and answers
	sb.WriteString("## Chi tiết bài thi\n\n")

	for _, record := range qs.QuestionHistory {
		resultIcon := "❌"
		resultText := "Sai"
		if record.IsCorrect {
			resultIcon = "✅"
			resultText = "Đúng"
		}

		sb.WriteString(fmt.Sprintf("### Câu %d %s\n\n", record.QuestionNumber, resultIcon))
		sb.WriteString(fmt.Sprintf("**Câu hỏi:** %s\n\n", record.Question))
		sb.WriteString(fmt.Sprintf("**Trả lời:** %s\n\n", record.StudentAnswer))
		sb.WriteString(fmt.Sprintf("**Kết quả:** %s (+%d điểm)\n\n", resultText, record.Points))
		if record.TeacherComment != "" {
			sb.WriteString(fmt.Sprintf("**Nhận xét:** %s\n\n", record.TeacherComment))
		}
		sb.WriteString("---\n\n")
	}

	// Current score (if exam in progress)
	if !qs.IsComplete && qs.CurrentQuestion > 0 {
		sb.WriteString("## Tiến độ hiện tại\n\n")
		sb.WriteString(fmt.Sprintf("- Đã hoàn thành: **%d/%d** câu\n", qs.CurrentQuestion, qs.TotalQuestions))
		sb.WriteString(fmt.Sprintf("- Điểm hiện tại: **%d** điểm\n", qs.CorrectAnswers))
		sb.WriteString(fmt.Sprintf("- Còn lại: **%d** câu\n\n", qs.TotalQuestions-qs.CurrentQuestion))
	}

	// Final result (if exam complete)
	if qs.IsComplete {
		percentage := float64(qs.CorrectAnswers) / float64(qs.TotalQuestions) * 100
		passed := qs.CorrectAnswers > 5
		grade := "CHƯA ĐẠT"
		gradeIcon := "🔴"
		if passed {
			grade = "ĐẠT"
			gradeIcon = "🟢"
		}

		sb.WriteString("## Kết quả cuối cùng\n\n")
		sb.WriteString(fmt.Sprintf("| Hạng mục | Kết quả |\n"))
		sb.WriteString(fmt.Sprintf("|----------|----------|\n"))
		sb.WriteString(fmt.Sprintf("| **Số câu đúng** | %d/%d |\n", qs.CorrectAnswers, qs.TotalQuestions))
		sb.WriteString(fmt.Sprintf("| **Số câu sai** | %d/%d |\n", qs.WrongAnswers, qs.TotalQuestions))
		sb.WriteString(fmt.Sprintf("| **Điểm số** | %.1f%% |\n", percentage))
		sb.WriteString(fmt.Sprintf("| **Xếp loại** | %s %s |\n", gradeIcon, grade))
		sb.WriteString("\n")

		// Teacher's final comment
		if teacherComment != "" {
			sb.WriteString("## Nhận xét của giáo viên\n\n")
			sb.WriteString(fmt.Sprintf("%s\n\n", teacherComment))
		}

		sb.WriteString("---\n\n")
		sb.WriteString(fmt.Sprintf("*Biên bản được tạo tự động lúc %s*\n", time.Now().Format("15:04:05 - 02/01/2006")))
	}

	return sb.String()
}

// WriteReportToFile writes the current report to the markdown file
func (qs *QuizState) WriteReportToFile(teacherComment string) error {
	qs.mu.RLock()
	reportPath := qs.ReportPath
	qs.mu.RUnlock()

	// Ensure directory exists
	dir := filepath.Dir(reportPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	content := qs.GenerateMarkdownReport(teacherComment)
	if err := os.WriteFile(reportPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	return nil
}

// GetStatus returns current quiz status
func (qs *QuizState) GetStatus() map[string]interface{} {
	qs.mu.RLock()
	defer qs.mu.RUnlock()

	return map[string]interface{}{
		"total_questions":    qs.TotalQuestions,
		"questions_asked":    qs.CurrentQuestion,
		"questions_remaining": qs.TotalQuestions - qs.CurrentQuestion,
		"correct_answers":    qs.CorrectAnswers,
		"wrong_answers":      qs.WrongAnswers,
		"current_score":      qs.CorrectAnswers,
		"is_complete":        qs.IsComplete,
	}
}

// RecordAnswer records an answer result with full details
func (qs *QuizState) RecordAnswer(questionNum int, question, studentAnswer string, isCorrect bool, teacherComment string) map[string]interface{} {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	// ✅ FIX: Hard limit - reject any question beyond TotalQuestions
	if qs.CurrentQuestion >= qs.TotalQuestions {
		return map[string]interface{}{
			"error":       "Kỳ thi đã hoàn thành! Không thể ghi thêm câu trả lời.",
			"is_complete": true,
			"final_score": qs.CorrectAnswers,
			"action":      "Vui lòng gọi GetFinalResult() và kết thúc với [KẾT THÚC THI]",
		}
	}

	// ✅ PHASE 3.6.1: Auto-infer question number if not provided (questionNum == 0)
	if questionNum == 0 || questionNum != qs.CurrentQuestion+1 {
		questionNum = qs.CurrentQuestion + 1
	}

	// Record the answer
	points := 0
	if isCorrect {
		points = 1
		qs.CorrectAnswers++
	} else {
		qs.WrongAnswers++
	}

	record := QuestionRecord{
		QuestionNumber: questionNum,
		Question:       question,
		StudentAnswer:  studentAnswer,
		IsCorrect:      isCorrect,
		Points:         points,
		TeacherComment: teacherComment,
	}
	qs.QuestionHistory = append(qs.QuestionHistory, record)
	qs.CurrentQuestion++

	// Check if complete
	if qs.CurrentQuestion >= qs.TotalQuestions {
		qs.IsComplete = true
	}

	return map[string]interface{}{
		"question_number":     questionNum,
		"question":            question,
		"student_answer":      studentAnswer,
		"is_correct":          isCorrect,
		"points_awarded":      points,
		"total_score":         qs.CorrectAnswers,
		"questions_remaining": qs.TotalQuestions - qs.CurrentQuestion,
		"is_complete":         qs.IsComplete,
	}
}

// GetFinalResult returns the final exam result
func (qs *QuizState) GetFinalResult() map[string]interface{} {
	qs.mu.RLock()
	defer qs.mu.RUnlock()

	passed := qs.CorrectAnswers > 5
	grade := "CHƯA ĐẠT"
	if passed {
		grade = "ĐẠT"
	}

	// Calculate percentage
	percentage := float64(qs.CorrectAnswers) / float64(qs.TotalQuestions) * 100

	return map[string]interface{}{
		"total_questions":  qs.TotalQuestions,
		"correct_answers":  qs.CorrectAnswers,
		"wrong_answers":    qs.WrongAnswers,
		"final_score":      qs.CorrectAnswers,
		"max_score":        qs.TotalQuestions,
		"percentage":       fmt.Sprintf("%.1f%%", percentage),
		"passed":           passed,
		"grade":            grade,
		"pass_threshold":   "> 5 điểm",
		"question_history": qs.QuestionHistory,
	}
}

// Reset resets the quiz state for a new exam
func (qs *QuizState) Reset() {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	qs.CurrentQuestion = 0
	qs.CorrectAnswers = 0
	qs.WrongAnswers = 0
	qs.QuestionHistory = make([]QuestionRecord, 0)
	qs.IsComplete = false
}

// CreateQuizTools creates the tools for the quiz exam
func CreateQuizTools(state *QuizState) map[string]*agenticcore.Tool {
	tools := make(map[string]*agenticcore.Tool)

	// Tool 1: GetQuizStatus - Get current quiz status
	tools["GetQuizStatus"] = &agenticcore.Tool{
		Name:        "GetQuizStatus",
		Description: "Lấy trạng thái hiện tại của kỳ thi: số câu đã hỏi, điểm hiện tại, còn bao nhiêu câu",
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Func: func(ctx context.Context, args map[string]interface{}) (string, error) {
			fmt.Fprintf(os.Stderr, "[TOOL ENTRY] GetQuizStatus() called\n")
			result := state.GetStatus()

			fmt.Printf("\n[TOOL] GetQuizStatus()\n")
			fmt.Printf("  Câu đã hỏi: %d/%d\n", result["questions_asked"], result["total_questions"])
			fmt.Printf("  Điểm hiện tại: %d\n", result["current_score"])
			fmt.Printf("  Còn lại: %d câu\n", result["questions_remaining"])
			fmt.Printf("  [DEBUG] state pointer: %p, CorrectAnswers: %d, CurrentQuestion: %d\n\n", state, state.CorrectAnswers, state.CurrentQuestion)
			fmt.Fprintf(os.Stderr, "[TOOL EXIT] GetQuizStatus() returning: questions_remaining=%d\n\n", result["questions_remaining"])

			jsonBytes, _ := json.Marshal(result)
			return string(jsonBytes), nil
		},
	}

	// Tool 2: RecordAnswer - Record answer result with full details
	tools["RecordAnswer"] = &agenticcore.Tool{
		Name:        "RecordAnswer",
		Description: "Ghi nhận kết quả câu trả lời của học sinh. SAU KHI GỌI TOOL NÀY, BẮT BUỘC PHẢI GỌI WriteExamReport để cập nhật biên bản.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"question_number": map[string]interface{}{
					"type":        "integer",
					"description": "Số thứ tự câu hỏi (1-10) - sẽ tự động suy ra nếu không cung cấp",
				},
				"question": map[string]interface{}{
					"type":        "string",
					"description": "Nội dung câu hỏi đã đặt",
				},
				"student_answer": map[string]interface{}{
					"type":        "string",
					"description": "Câu trả lời của học sinh",
				},
				"is_correct": map[string]interface{}{
					"type":        "boolean",
					"description": "true nếu học sinh trả lời đúng, false nếu sai",
				},
				"teacher_comment": map[string]interface{}{
					"type":        "string",
					"description": "Nhận xét ngắn gọn của giáo viên về câu trả lời (tùy chọn)",
				},
			},
			"required": []string{"question", "student_answer", "is_correct"},
		},
		Func: func(ctx context.Context, args map[string]interface{}) (string, error) {
			fmt.Fprintf(os.Stderr, "[TOOL ENTRY] RecordAnswer() called with args: %v\n", args)

			// ✅ PHASE 3.6: Auto-infer question_number from current state
			var questionNum int

			if qn, exists := args["question_number"]; exists && qn != nil {
				// LLM cung cấp question_number - sử dụng giá trị này
				switch v := qn.(type) {
				case float64:
					questionNum = int(v)
				case int64:
					questionNum = int(v)
				case int:
					questionNum = v
				default:
					// Fallback: Suy ra từ state (KHÔNG DÙNG - sẽ auto-infer trong RecordAnswer)
					questionNum = 0
				}
			} else {
				// LLM không cung cấp - tự động suy ra từ trạng thái hiện tại trong RecordAnswer
				questionNum = 0
			}

			// Parse question
			question, _ := args["question"].(string)

			// Parse student_answer (handle both string and numeric types)
			var studentAnswer string
			switch v := args["student_answer"].(type) {
			case string:
				studentAnswer = v
			case float64:
				studentAnswer = fmt.Sprintf("%v", v)
			case int64:
				studentAnswer = fmt.Sprintf("%d", v)
			case int:
				studentAnswer = fmt.Sprintf("%d", v)
			default:
				studentAnswer = fmt.Sprintf("%v", v)
			}

			// ✅ PHASE 3.6: Auto-detect is_correct (fallback to true if not provided)
			isCorrect := true  // Default: assume answer is correct
			if ic, exists := args["is_correct"]; exists && ic != nil {
				if b, ok := ic.(bool); ok {
					isCorrect = b
				}
			}

			// Parse teacher_comment (optional)
			teacherComment, _ := args["teacher_comment"].(string)

			fmt.Fprintf(os.Stderr, "[TOOL DEBUG] Before RecordAnswer: CurrentQuestion=%d, TotalQuestions=%d\n", state.CurrentQuestion, state.TotalQuestions)
			result := state.RecordAnswer(questionNum, question, studentAnswer, isCorrect, teacherComment)
			fmt.Fprintf(os.Stderr, "[TOOL DEBUG] After RecordAnswer: CurrentQuestion=%d, is_complete=%v\n", state.CurrentQuestion, result["is_complete"])

			// Check for error
			if _, hasError := result["error"]; hasError {
				fmt.Printf("\n[TOOL ERROR] RecordAnswer: %s\n\n", result["error"])
				fmt.Fprintf(os.Stderr, "[TOOL ERROR] RecordAnswer returned error: %v\n", result["error"])
				jsonBytes, _ := json.Marshal(result)
				return string(jsonBytes), nil
			}

			correctStr := "SAI"
			if isCorrect {
				correctStr = "ĐÚNG"
			}
			fmt.Printf("\n[TOOL] RecordAnswer(question=%d, is_correct=%v)\n", result["question_number"], isCorrect)
			fmt.Printf("  Câu hỏi: %s\n", question)
			fmt.Printf("  Trả lời: %s\n", studentAnswer)
			fmt.Printf("  Kết quả: %s (+%d điểm)\n", correctStr, result["points_awarded"])
			fmt.Printf("  Tổng điểm: %d\n", result["total_score"])
			fmt.Printf("  Còn lại: %d câu\n", result["questions_remaining"])
			fmt.Printf("  [DEBUG] state pointer: %p, CorrectAnswers: %d, CurrentQuestion: %d\n\n", state, state.CorrectAnswers, state.CurrentQuestion)

			// ✅ AUTO-SAVE: Write report after each answer to ensure file is updated
			if err := state.WriteReportToFile(""); err != nil {
				fmt.Printf("  [Auto-save] Lỗi lưu biên bản: %v\n", err)
				fmt.Fprintf(os.Stderr, "[TOOL ERROR] WriteReportToFile failed: %v\n", err)
			}

			fmt.Fprintf(os.Stderr, "[TOOL EXIT] RecordAnswer() returning: is_complete=%v, questions_remaining=%d\n\n", result["is_complete"], result["questions_remaining"])
			jsonBytes, _ := json.Marshal(result)
			return string(jsonBytes), nil
		},
	}

	// Tool 3: GetFinalResult - Get final exam result
	tools["GetFinalResult"] = &agenticcore.Tool{
		Name:        "GetFinalResult",
		Description: "Lấy kết quả cuối cùng của kỳ thi (chỉ gọi khi đã đủ 10 câu). SAU KHI GỌI TOOL NÀY, BẮT BUỘC PHẢI GỌI WriteExamReport với nhận xét tổng kết.",
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Func: func(ctx context.Context, args map[string]interface{}) (string, error) {
			fmt.Fprintf(os.Stderr, "[TOOL ENTRY] GetFinalResult() called\n")
			result := state.GetFinalResult()

			fmt.Printf("\n[TOOL] GetFinalResult()\n")
			fmt.Printf("  ========== KẾT QUẢ THI ==========\n")
			fmt.Printf("  Tổng số câu: %d\n", result["total_questions"])
			fmt.Printf("  Số câu đúng: %d\n", result["correct_answers"])
			fmt.Printf("  Số câu sai: %d\n", result["wrong_answers"])
			fmt.Printf("  Điểm số: %d/%d (%s)\n", result["final_score"], result["max_score"], result["percentage"])
			fmt.Printf("  Kết quả: %s\n", result["grade"])
			fmt.Printf("  ==================================\n\n")
			fmt.Fprintf(os.Stderr, "[TOOL EXIT] GetFinalResult() returning: grade=%v\n\n", result["grade"])

			jsonBytes, _ := json.Marshal(result)
			return string(jsonBytes), nil
		},
	}

	// Tool 4: WriteExamReport - Write/update exam report to markdown file
	tools["WriteExamReport"] = &agenticcore.Tool{
		Name:        "WriteExamReport",
		Description: "Ghi/cập nhật biên bản thi vấn đáp ra file markdown. PHẢI GỌI SAU MỖI LẦN RecordAnswer và sau GetFinalResult.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"teacher_final_comment": map[string]interface{}{
					"type":        "string",
					"description": "Nhận xét tổng kết của giáo viên (chỉ cần khi kết thúc bài thi)",
				},
			},
		},
		Func: func(ctx context.Context, args map[string]interface{}) (string, error) {
			fmt.Fprintf(os.Stderr, "[TOOL ENTRY] WriteExamReport() called\n")

			// Parse teacher_final_comment (optional)
			teacherComment, _ := args["teacher_final_comment"].(string)

			// Write report to file
			err := state.WriteReportToFile(teacherComment)
			if err != nil {
				fmt.Printf("\n[TOOL ERROR] WriteExamReport: %v\n\n", err)
				fmt.Fprintf(os.Stderr, "[TOOL ERROR] WriteExamReport failed: %v\n", err)
				return fmt.Sprintf(`{"error": "%v"}`, err), nil
			}

			state.mu.RLock()
			reportPath := state.ReportPath
			isComplete := state.IsComplete
			currentQ := state.CurrentQuestion
			state.mu.RUnlock()

			status := "đang thi"
			if isComplete {
				status = "hoàn thành"
			}

			fmt.Printf("\n[TOOL] WriteExamReport()\n")
			fmt.Printf("  File: %s\n", reportPath)
			fmt.Printf("  Trạng thái: %s (%d câu)\n\n", status, currentQ)
			fmt.Fprintf(os.Stderr, "[TOOL EXIT] WriteExamReport() success: status=%v, questions=%d\n\n", status, currentQ)

			result := map[string]interface{}{
				"success":     true,
				"report_path": reportPath,
				"status":      status,
				"questions":   currentQ,
			}
			jsonBytes, _ := json.Marshal(result)
			return string(jsonBytes), nil
		},
	}

	// Tool 5: SetExamInfo - Set student name and exam topic
	tools["SetExamInfo"] = &agenticcore.Tool{
		Name:        "SetExamInfo",
		Description: "Đặt thông tin kỳ thi: tên học sinh và chủ đề thi. Nên gọi ở đầu kỳ thi.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"student_name": map[string]interface{}{
					"type":        "string",
					"description": "Tên của học sinh",
				},
				"exam_topic": map[string]interface{}{
					"type":        "string",
					"description": "Chủ đề của kỳ thi",
				},
			},
		},
		Func: func(ctx context.Context, args map[string]interface{}) (string, error) {
			studentName, _ := args["student_name"].(string)
			examTopic, _ := args["exam_topic"].(string)

			state.SetStudentInfo(studentName, examTopic)

			fmt.Printf("\n[TOOL] SetExamInfo()\n")
			if studentName != "" {
				fmt.Printf("  Học sinh: %s\n", studentName)
			}
			if examTopic != "" {
				fmt.Printf("  Chủ đề: %s\n", examTopic)
			}
			fmt.Println()

			result := map[string]interface{}{
				"success":      true,
				"student_name": studentName,
				"exam_topic":   examTopic,
			}
			jsonBytes, _ := json.Marshal(result)
			return string(jsonBytes), nil
		},
	}

	return tools
}
