package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dariubs/altern/app/config"
	"github.com/dariubs/altern/app/database"
	"github.com/dariubs/altern/app/models"
)

func main() {
	force := len(os.Args) > 1 && os.Args[1] == "--force"

	if err := config.Load(); err != nil {
		log.Fatal(err)
	}
	database.InitDB()
	db := database.DB

	if !force {
		var count int64
		db.Model(&models.Tool{}).Count(&count)
		if count > 0 {
			log.Fatalf("%d tools already exist; run with FORCE=1 to wipe and re-seed", count)
		}
	} else {
		log.Println("wiping existing data…")
		db.Exec("TRUNCATE users, categories, tools, lists, list_tools, tool_categories RESTART IDENTITY CASCADE")
	}

	// ── Categories ────────────────────────────────────────────────────────────

	type catDef struct {
		Name     string
		Subtitle string
		Parent   string
	}

	catDefs := []catDef{
		// top-level (10)
		{"Writing & Content", "Tools for writing, editing, and content creation", ""},
		{"Image Generation", "AI tools that create images from text or other inputs", ""},
		{"Code & Development", "AI coding assistants, editors, and dev tools", ""},
		{"Video & Animation", "AI tools for creating and editing video content", ""},
		{"Audio & Music", "Voice synthesis, music generation, and audio editing", ""},
		{"Data & Analytics", "AI-powered data analysis and business intelligence", ""},
		{"Productivity", "AI tools to help you work and think faster", ""},
		{"Search & Research", "AI-enhanced search engines and research assistants", ""},
		{"Chatbots & Assistants", "General-purpose conversational AI", ""},
		{"Design & UX", "AI tools for UI, graphic, and product design", ""},
		// subcategories (10)
		{"Copywriting", "AI for marketing copy and brand content", "Writing & Content"},
		{"Translation", "AI-powered language translation", "Writing & Content"},
		{"Art Generation", "Fine art and creative image generation", "Image Generation"},
		{"Photo Editing", "AI-enhanced photo retouching and manipulation", "Image Generation"},
		{"Code Completion", "AI autocompletion and in-editor suggestions", "Code & Development"},
		{"Code Review", "AI tools for reviewing and improving code quality", "Code & Development"},
		{"Text-to-Video", "Generate video clips from text descriptions", "Video & Animation"},
		{"Voice & Speech", "Text-to-speech, transcription, and voice cloning", "Audio & Music"},
		{"Business Intelligence", "AI dashboards, predictions, and data insights", "Data & Analytics"},
		{"Note Taking", "AI-assisted notes, memory, and knowledge management", "Productivity"},
	}

	catMap := make(map[string]*models.Category, len(catDefs))

	for _, def := range catDefs {
		if def.Parent != "" {
			continue
		}
		c := models.Category{Name: def.Name, Slug: toSlug(def.Name), Subtitle: def.Subtitle}
		if err := db.Create(&c).Error; err != nil {
			log.Fatalf("create category %q: %v", def.Name, err)
		}
		catMap[def.Name] = &c
	}
	for _, def := range catDefs {
		if def.Parent == "" {
			continue
		}
		parent := catMap[def.Parent]
		c := models.Category{Name: def.Name, Slug: toSlug(def.Name), Subtitle: def.Subtitle, ParentID: &parent.ID}
		if err := db.Create(&c).Error; err != nil {
			log.Fatalf("create subcategory %q: %v", def.Name, err)
		}
		catMap[def.Name] = &c
	}
	fmt.Printf("  categories: %d\n", len(catDefs))

	// ── Tools ─────────────────────────────────────────────────────────────────

	type toolDef struct {
		Name       string
		Subtitle   string
		Categories []string
	}

	toolDefs := []toolDef{
		// Chatbots & Assistants (5)
		{"ChatGPT", "Conversational AI by OpenAI", []string{"Chatbots & Assistants"}},
		{"Claude", "AI assistant focused on safety and helpfulness", []string{"Chatbots & Assistants"}},
		{"Gemini", "Google's multimodal AI assistant", []string{"Chatbots & Assistants"}},
		{"Mistral Chat", "Open-source conversational AI", []string{"Chatbots & Assistants"}},
		{"Pi", "Personal AI companion by Inflection", []string{"Chatbots & Assistants"}},
		// Image Generation (6)
		{"Midjourney", "High-quality AI art from text prompts", []string{"Image Generation", "Art Generation"}},
		{"DALL-E 3", "OpenAI's latest image generation model", []string{"Image Generation"}},
		{"Stable Diffusion", "Open-source AI image generation", []string{"Image Generation", "Art Generation"}},
		{"Adobe Firefly", "Creative AI built into Adobe's suite", []string{"Image Generation", "Photo Editing"}},
		{"Leonardo.ai", "AI image generation for game assets and art", []string{"Image Generation", "Art Generation"}},
		{"Ideogram", "Text-accurate AI image generation", []string{"Image Generation"}},
		// Code & Development (6)
		{"GitHub Copilot", "AI pair programmer in your editor", []string{"Code & Development", "Code Completion"}},
		{"Cursor", "AI-first code editor", []string{"Code & Development", "Code Completion"}},
		{"Tabnine", "AI code completion for any IDE", []string{"Code & Development", "Code Completion"}},
		{"Codeium", "Free AI code completion and chat", []string{"Code & Development", "Code Completion"}},
		{"Phind", "AI search engine for developers", []string{"Code & Development", "Search & Research"}},
		{"Replit AI", "AI-powered collaborative coding environment", []string{"Code & Development"}},
		// Writing & Content (5)
		{"Jasper", "AI writing platform for marketing teams", []string{"Writing & Content", "Copywriting"}},
		{"Copy.ai", "AI copywriting and content tool", []string{"Writing & Content", "Copywriting"}},
		{"Writesonic", "AI content creation for blogs and ads", []string{"Writing & Content", "Copywriting"}},
		{"Grammarly", "AI writing assistant for clarity and correctness", []string{"Writing & Content"}},
		{"DeepL", "AI-powered translation with nuance", []string{"Writing & Content", "Translation"}},
		// Video & Animation (5)
		{"Runway", "AI video creation and editing suite", []string{"Video & Animation", "Text-to-Video"}},
		{"Pika", "Generate and edit video with AI", []string{"Video & Animation", "Text-to-Video"}},
		{"HeyGen", "AI video avatar and dubbing platform", []string{"Video & Animation"}},
		{"Synthesia", "Create AI presenter videos at scale", []string{"Video & Animation"}},
		{"D-ID", "Animate photos with AI-generated speech", []string{"Video & Animation"}},
		// Audio & Music (6)
		{"ElevenLabs", "Realistic AI voice synthesis and cloning", []string{"Audio & Music", "Voice & Speech"}},
		{"Suno", "Generate full songs with AI", []string{"Audio & Music"}},
		{"Udio", "AI music creation from text descriptions", []string{"Audio & Music"}},
		{"Whisper", "OpenAI's open-source speech recognition", []string{"Audio & Music", "Voice & Speech"}},
		{"Mubert", "AI-generated royalty-free music streams", []string{"Audio & Music"}},
		{"Descript", "AI-powered audio and video editing", []string{"Audio & Music", "Video & Animation"}},
		// Data & Analytics (5)
		{"Obviously AI", "No-code AI predictions for business data", []string{"Data & Analytics", "Business Intelligence"}},
		{"Akkio", "AI analytics platform for growth teams", []string{"Data & Analytics", "Business Intelligence"}},
		{"Airtable AI", "Database and workflow automation with AI", []string{"Data & Analytics", "Productivity"}},
		{"Julius", "Chat with your data using AI", []string{"Data & Analytics"}},
		{"MonkeyLearn", "No-code text analysis and classification", []string{"Data & Analytics"}},
		// Productivity (5)
		{"Notion AI", "AI writing and summarization inside Notion", []string{"Productivity", "Note Taking"}},
		{"Mem", "AI-powered self-organizing notes", []string{"Productivity", "Note Taking"}},
		{"Otter.ai", "AI meeting transcription and summaries", []string{"Productivity", "Voice & Speech"}},
		{"Motion", "AI calendar, task, and project manager", []string{"Productivity"}},
		{"Reclaim.ai", "AI scheduling assistant for busy teams", []string{"Productivity"}},
		// Search & Research (2)
		{"Perplexity", "AI-powered search with cited answers", []string{"Search & Research"}},
		{"You.com", "AI search engine and assistant", []string{"Search & Research", "Chatbots & Assistants"}},
		// Design & UX (5)
		{"Framer AI", "AI-powered website and landing page builder", []string{"Design & UX"}},
		{"Uizard", "Turn sketches into UI designs with AI", []string{"Design & UX"}},
		{"v0", "Generate UI components with AI by Vercel", []string{"Design & UX", "Code & Development"}},
		{"Canva AI", "AI design tools inside Canva", []string{"Design & UX", "Image Generation"}},
		{"Galileo AI", "Generate editable UI designs from text", []string{"Design & UX"}},
	}

	toolObjs := make([]*models.Tool, 0, len(toolDefs))
	for _, def := range toolDefs {
		t := models.Tool{Name: def.Name, Slug: toSlug(def.Name), Subtitle: def.Subtitle}
		if err := db.Create(&t).Error; err != nil {
			log.Fatalf("create tool %q: %v", def.Name, err)
		}
		cats := make([]models.Category, 0, len(def.Categories))
		for _, cn := range def.Categories {
			if c, ok := catMap[cn]; ok {
				cats = append(cats, *c)
			}
		}
		if len(cats) > 0 {
			if err := db.Model(&t).Association("Categories").Replace(cats); err != nil {
				log.Fatalf("assign categories for %q: %v", def.Name, err)
			}
		}
		toolObjs = append(toolObjs, &t)
	}
	fmt.Printf("  tools: %d\n", len(toolDefs))

	// ── Users ─────────────────────────────────────────────────────────────────

	type userDef struct {
		Name     string
		Username string
		Email    string
		GHLogin  string
	}

	userDefs := []userDef{
		{"Alice Morgan", "alice", "alice@example.com", "alice-morgan"},
		{"Bob Chen", "bob", "bob@example.com", "bob-chen"},
		{"Carol Davis", "carol", "carol@example.com", "carol-davis"},
		{"Dave Kim", "dave", "dave@example.com", "dave-kim"},
		{"Eve Nakamura", "eve", "eve@example.com", "eve-nakamura"},
		{"Frank Osei", "frank", "frank@example.com", "frank-osei"},
		{"Grace Lim", "grace", "grace@example.com", "grace-lim"},
		{"Henry Torres", "henry", "henry@example.com", "henry-torres"},
		{"Iris Petrov", "iris", "iris@example.com", "iris-petrov"},
		{"Jack Williams", "jack", "jack@example.com", "jack-williams"},
	}

	for i, def := range userDefs {
		u := models.User{
			Name:        def.Name,
			Username:    def.Username,
			Email:       def.Email,
			GitHubLogin: def.GHLogin,
			GitHubID:    fmt.Sprintf("seed_%s_%03d", def.Username, i+1),
		}
		if err := db.Create(&u).Error; err != nil {
			log.Fatalf("create user %q: %v", def.Username, err)
		}
	}
	fmt.Printf("  users: %d\n", len(userDefs))

	// ── Lists ─────────────────────────────────────────────────────────────────

	type listDef struct {
		Name     string
		Subtitle string
		ToolIdxs []int
	}

	listDefs := []listDef{
		{"Best AI Chatbots", "The top conversational AI assistants you can use today", []int{0, 1, 2, 3, 4}},
		{"Top Image Generators", "AI tools that turn text into stunning visuals", []int{5, 6, 7, 8, 9, 10}},
		{"AI Coding Assistants", "Write, complete, and review code faster with AI", []int{11, 12, 13, 14, 15, 16}},
		{"AI Writing Tools", "Write better content with AI assistance", []int{17, 18, 19, 20, 21}},
		{"AI Video Generators", "Create and edit video entirely with AI", []int{22, 23, 24, 25, 26}},
		{"AI Audio & Music", "Voice, music, and sound tools powered by AI", []int{27, 28, 29, 30, 31, 32}},
		{"AI Productivity Apps", "Work smarter with AI-powered productivity tools", []int{38, 39, 40, 41, 42}},
		{"AI Data Tools", "Analyze and visualize data using AI", []int{33, 34, 35, 36, 37}},
		{"Free AI Tools", "Powerful AI tools you can start using for free", []int{1, 4, 7, 14, 30, 39, 43}},
		{"Getting Started with AI", "The best AI tools for beginners", []int{0, 1, 5, 11, 17, 38, 43}},
	}

	for _, def := range listDefs {
		l := models.List{Name: def.Name, Subtitle: def.Subtitle}
		if err := db.Create(&l).Error; err != nil {
			log.Fatalf("create list %q: %v", def.Name, err)
		}
		for sort, ti := range def.ToolIdxs {
			lt := models.ListTool{ListID: l.ID, ToolID: toolObjs[ti].ID, Sort: sort}
			if err := db.Create(&lt).Error; err != nil {
				log.Fatalf("create list_tool: %v", err)
			}
		}
	}
	fmt.Printf("  lists: %d\n", len(listDefs))

	fmt.Println("seed complete.")
}

func toSlug(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	prev := '-'
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prev = r
		} else if prev != '-' {
			b.WriteByte('-')
			prev = '-'
		}
	}
	return strings.Trim(b.String(), "-")
}
