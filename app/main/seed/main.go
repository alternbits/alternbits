package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/dariubs/altern/app/config"
	"github.com/dariubs/altern/app/database"
	"github.com/dariubs/altern/app/models"
	"gorm.io/gorm"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatal(err)
	}
	database.InitDB()
	db := database.DB

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
		slug := toSlug(def.Name)
		var c models.Category
		db.Where("slug = ?", slug).First(&c)
		c.Name = def.Name
		c.Slug = slug
		c.Subtitle = def.Subtitle
		c.ParentID = nil
		if err := upsert(db, c.ID, &c); err != nil {
			log.Fatalf("upsert category %q: %v", def.Name, err)
		}
		catMap[def.Name] = &c
	}
	for _, def := range catDefs {
		if def.Parent == "" {
			continue
		}
		parent := catMap[def.Parent]
		slug := toSlug(def.Name)
		var c models.Category
		db.Where("slug = ?", slug).First(&c)
		c.Name = def.Name
		c.Slug = slug
		c.Subtitle = def.Subtitle
		c.ParentID = &parent.ID
		if err := upsert(db, c.ID, &c); err != nil {
			log.Fatalf("upsert subcategory %q: %v", def.Name, err)
		}
		catMap[def.Name] = &c
	}
	fmt.Printf("  categories: %d\n", len(catDefs))

	// ── Genera ────────────────────────────────────────────────────────────────

	type genusDef struct {
		Name     string
		Subtitle string
	}

	genusDefs := []genusDef{
		{"Web App", "Runs in a browser with no install required"},
		{"Mobile App", "Native iOS or Android application"},
		{"Desktop App", "Installable app for macOS, Windows, or Linux"},
		{"Browser Extension", "Plugin that runs inside your browser"},
		{"API", "Programmatic interface for developers"},
		{"AI Model", "A standalone model with published weights or endpoints"},
		{"Open Source", "Source code is publicly available"},
		{"Python Library", "Installable Python package"},
		{"CLI Tool", "Command-line interface tool"},
		{"AI Agent", "Autonomous agent that can take multi-step actions"},
	}

	genusMap := make(map[string]*models.Genus, len(genusDefs))
	for _, def := range genusDefs {
		slug := toSlug(def.Name)
		var g models.Genus
		db.Where("slug = ?", slug).First(&g)
		g.Name = def.Name
		g.Slug = slug
		g.Subtitle = def.Subtitle
		if err := upsert(db, g.ID, &g); err != nil {
			log.Fatalf("upsert genus %q: %v", def.Name, err)
		}
		genusMap[def.Name] = &g
	}
	fmt.Printf("  genera: %d\n", len(genusDefs))

	// ── Tools ─────────────────────────────────────────────────────────────────

	type toolDef struct {
		Name       string
		Subtitle   string
		Categories []string
		Genera     []string
	}

	toolDefs := []toolDef{
		// Chatbots & Assistants (5)
		{"ChatGPT", "Conversational AI by OpenAI", []string{"Chatbots & Assistants"}, []string{"Web App", "Mobile App", "API"}},
		{"Claude", "AI assistant focused on safety and helpfulness", []string{"Chatbots & Assistants"}, []string{"Web App", "Mobile App", "API"}},
		{"Gemini", "Google's multimodal AI assistant", []string{"Chatbots & Assistants"}, []string{"Web App", "Mobile App", "API"}},
		{"Mistral Chat", "Open-source conversational AI", []string{"Chatbots & Assistants"}, []string{"Web App", "API", "AI Model", "Open Source"}},
		{"Pi", "Personal AI companion by Inflection", []string{"Chatbots & Assistants"}, []string{"Web App", "Mobile App"}},
		// Image Generation (6)
		{"Midjourney", "High-quality AI art from text prompts", []string{"Image Generation", "Art Generation"}, []string{"Web App"}},
		{"DALL-E 3", "OpenAI's latest image generation model", []string{"Image Generation"}, []string{"API", "AI Model"}},
		{"Stable Diffusion", "Open-source AI image generation", []string{"Image Generation", "Art Generation"}, []string{"AI Model", "Open Source", "Python Library"}},
		{"Adobe Firefly", "Creative AI built into Adobe's suite", []string{"Image Generation", "Photo Editing"}, []string{"Web App", "Desktop App"}},
		{"Leonardo.ai", "AI image generation for game assets and art", []string{"Image Generation", "Art Generation"}, []string{"Web App", "API"}},
		{"Ideogram", "Text-accurate AI image generation", []string{"Image Generation"}, []string{"Web App", "API"}},
		// Code & Development (6)
		{"GitHub Copilot", "AI pair programmer in your editor", []string{"Code & Development", "Code Completion"}, []string{"Browser Extension", "API"}},
		{"Cursor", "AI-first code editor", []string{"Code & Development", "Code Completion"}, []string{"Desktop App"}},
		{"Tabnine", "AI code completion for any IDE", []string{"Code & Development", "Code Completion"}, []string{"Desktop App", "Browser Extension"}},
		{"Codeium", "Free AI code completion and chat", []string{"Code & Development", "Code Completion"}, []string{"Web App", "Desktop App", "Browser Extension"}},
		{"Phind", "AI search engine for developers", []string{"Code & Development", "Search & Research"}, []string{"Web App"}},
		{"Replit AI", "AI-powered collaborative coding environment", []string{"Code & Development"}, []string{"Web App", "Mobile App"}},
		// Writing & Content (5)
		{"Jasper", "AI writing platform for marketing teams", []string{"Writing & Content", "Copywriting"}, []string{"Web App", "Browser Extension"}},
		{"Copy.ai", "AI copywriting and content tool", []string{"Writing & Content", "Copywriting"}, []string{"Web App"}},
		{"Writesonic", "AI content creation for blogs and ads", []string{"Writing & Content", "Copywriting"}, []string{"Web App", "API"}},
		{"Grammarly", "AI writing assistant for clarity and correctness", []string{"Writing & Content"}, []string{"Web App", "Desktop App", "Browser Extension"}},
		{"DeepL", "AI-powered translation with nuance", []string{"Writing & Content", "Translation"}, []string{"Web App", "Desktop App", "API"}},
		// Video & Animation (5)
		{"Runway", "AI video creation and editing suite", []string{"Video & Animation", "Text-to-Video"}, []string{"Web App", "API"}},
		{"Pika", "Generate and edit video with AI", []string{"Video & Animation", "Text-to-Video"}, []string{"Web App", "Mobile App"}},
		{"HeyGen", "AI video avatar and dubbing platform", []string{"Video & Animation"}, []string{"Web App", "API"}},
		{"Synthesia", "Create AI presenter videos at scale", []string{"Video & Animation"}, []string{"Web App", "API"}},
		{"D-ID", "Animate photos with AI-generated speech", []string{"Video & Animation"}, []string{"Web App", "API"}},
		// Audio & Music (6)
		{"ElevenLabs", "Realistic AI voice synthesis and cloning", []string{"Audio & Music", "Voice & Speech"}, []string{"Web App", "Mobile App", "API"}},
		{"Suno", "Generate full songs with AI", []string{"Audio & Music"}, []string{"Web App", "Mobile App"}},
		{"Udio", "AI music creation from text descriptions", []string{"Audio & Music"}, []string{"Web App"}},
		{"Whisper", "OpenAI's open-source speech recognition", []string{"Audio & Music", "Voice & Speech"}, []string{"AI Model", "Open Source", "Python Library", "CLI Tool"}},
		{"Mubert", "AI-generated royalty-free music streams", []string{"Audio & Music"}, []string{"Web App", "API"}},
		{"Descript", "AI-powered audio and video editing", []string{"Audio & Music", "Video & Animation"}, []string{"Desktop App", "Web App"}},
		// Data & Analytics (5)
		{"Obviously AI", "No-code AI predictions for business data", []string{"Data & Analytics", "Business Intelligence"}, []string{"Web App"}},
		{"Akkio", "AI analytics platform for growth teams", []string{"Data & Analytics", "Business Intelligence"}, []string{"Web App"}},
		{"Airtable AI", "Database and workflow automation with AI", []string{"Data & Analytics", "Productivity"}, []string{"Web App", "Mobile App"}},
		{"Julius", "Chat with your data using AI", []string{"Data & Analytics"}, []string{"Web App"}},
		{"MonkeyLearn", "No-code text analysis and classification", []string{"Data & Analytics"}, []string{"Web App", "API"}},
		// Productivity (5)
		{"Notion AI", "AI writing and summarization inside Notion", []string{"Productivity", "Note Taking"}, []string{"Web App", "Desktop App", "Mobile App"}},
		{"Mem", "AI-powered self-organizing notes", []string{"Productivity", "Note Taking"}, []string{"Web App", "Mobile App"}},
		{"Otter.ai", "AI meeting transcription and summaries", []string{"Productivity", "Voice & Speech"}, []string{"Web App", "Mobile App"}},
		{"Motion", "AI calendar, task, and project manager", []string{"Productivity"}, []string{"Web App", "Mobile App"}},
		{"Reclaim.ai", "AI scheduling assistant for busy teams", []string{"Productivity"}, []string{"Web App", "Browser Extension"}},
		// Search & Research (2)
		{"Perplexity", "AI-powered search with cited answers", []string{"Search & Research"}, []string{"Web App", "Mobile App", "API"}},
		{"You.com", "AI search engine and assistant", []string{"Search & Research", "Chatbots & Assistants"}, []string{"Web App"}},
		// Design & UX (5)
		{"Framer AI", "AI-powered website and landing page builder", []string{"Design & UX"}, []string{"Web App"}},
		{"Uizard", "Turn sketches into UI designs with AI", []string{"Design & UX"}, []string{"Web App"}},
		{"v0", "Generate UI components with AI by Vercel", []string{"Design & UX", "Code & Development"}, []string{"Web App", "API"}},
		{"Canva AI", "AI design tools inside Canva", []string{"Design & UX", "Image Generation"}, []string{"Web App", "Mobile App", "Desktop App"}},
		{"Galileo AI", "Generate editable UI designs from text", []string{"Design & UX"}, []string{"Web App"}},
	}

	toolObjs := make([]*models.Tool, 0, len(toolDefs))
	for _, def := range toolDefs {
		slug := toSlug(def.Name)
		var t models.Tool
		db.Where("slug = ?", slug).First(&t)
		t.Name = def.Name
		t.Slug = slug
		t.Subtitle = def.Subtitle
		if err := upsert(db, t.ID, &t); err != nil {
			log.Fatalf("upsert tool %q: %v", def.Name, err)
		}
		cats := make([]models.Category, 0, len(def.Categories))
		for _, cn := range def.Categories {
			if c, ok := catMap[cn]; ok {
				cats = append(cats, *c)
			}
		}
		if err := db.Model(&t).Association("Categories").Replace(cats); err != nil {
			log.Fatalf("assign categories for %q: %v", def.Name, err)
		}
		genera := make([]models.Genus, 0, len(def.Genera))
		for _, gn := range def.Genera {
			if g, ok := genusMap[gn]; ok {
				genera = append(genera, *g)
			}
		}
		if err := db.Model(&t).Association("Genera").Replace(genera); err != nil {
			log.Fatalf("assign genera for %q: %v", def.Name, err)
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
		var u models.User
		db.Where("username = ?", def.Username).First(&u)
		u.Name = def.Name
		u.Username = def.Username
		u.Email = def.Email
		u.GitHubLogin = def.GHLogin
		u.GitHubID = fmt.Sprintf("seed_%s_%03d", def.Username, i+1)
		if err := upsert(db, u.ID, &u); err != nil {
			log.Fatalf("upsert user %q: %v", def.Username, err)
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
		var l models.List
		db.Where("name = ?", def.Name).First(&l)
		l.Name = def.Name
		l.Subtitle = def.Subtitle
		if err := upsert(db, l.ID, &l); err != nil {
			log.Fatalf("upsert list %q: %v", def.Name, err)
		}
		db.Where("list_id = ?", l.ID).Delete(&models.ListTool{})
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

func upsert(db *gorm.DB, id uint, v any) error {
	if id == 0 {
		return db.Create(v).Error
	}
	return db.Save(v).Error
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
