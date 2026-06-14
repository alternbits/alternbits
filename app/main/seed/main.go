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
		{"Writing & Content", "AIs for writing, editing, and content creation", ""},
		{"Image Generation", "AIs that create images from text or other inputs", ""},
		{"Code & Development", "AI coding assistants, editors, and development AIs", ""},
		{"Video & Animation", "AIs for creating and editing video content", ""},
		{"Audio & Music", "Voice synthesis, music generation, and audio editing", ""},
		{"Data & Analytics", "AI-powered data analysis and business intelligence", ""},
		{"Productivity", "AIs to help you work and think faster", ""},
		{"Search & Research", "AI-enhanced search engines and research assistants", ""},
		{"Chatbots & Assistants", "General-purpose conversational AI", ""},
		{"Design & UX", "AIs for UI, graphic, and product design", ""},
		// subcategories (10)
		{"Copywriting", "AI for marketing copy and brand content", "Writing & Content"},
		{"Translation", "AI-powered language translation", "Writing & Content"},
		{"Art Generation", "Fine art and creative image generation", "Image Generation"},
		{"Photo Editing", "AI-enhanced photo retouching and manipulation", "Image Generation"},
		{"Code Completion", "AI autocompletion and in-editor suggestions", "Code & Development"},
		{"Code Review", "AIs for reviewing and improving code quality", "Code & Development"},
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

	// ── AIs ───────────────────────────────────────────────────────────────────

	type aiDef struct {
		Name        string
		Subtitle    string
		Description string
		Website     string
		HaveFree    bool
		Categories  []string
		Genera      []string
	}

	aiDefs := []aiDef{
		// Chatbots & Assistants (5)
		{
			"ChatGPT", "Conversational AI by OpenAI",
			"ChatGPT is OpenAI's flagship conversational AI, capable of writing, reasoning, coding, and analysis.\n\n## Key features\n\n- **GPT-4o** multimodal model with vision, voice, and file support\n- **Custom GPTs** — build and share tailored assistants\n- **Code Interpreter** for data analysis and chart generation\n- **Plugins and browsing** for real-time information\n- Available on web, iOS, Android, and via API\n\nChatGPT set the standard for modern AI assistants and remains the most widely used AI chatbot in the world.",
			"https://chatgpt.com", true,
			[]string{"Chatbots & Assistants"}, []string{"Web App", "Mobile App", "API"},
		},
		{
			"Claude", "AI assistant focused on safety and helpfulness",
			"Claude is Anthropic's AI assistant, designed with a strong emphasis on safety, honesty, and helpfulness.\n\n## Key features\n\n- **200K context window** — handle entire codebases or long documents\n- **Extended thinking** for step-by-step reasoning on complex tasks\n- **Projects** to keep context and files organized across sessions\n- Available on web, iOS, Android, and via API\n\nClaude is particularly strong at long-form writing, nuanced reasoning, and following complex instructions faithfully.",
			"https://claude.ai", true,
			[]string{"Chatbots & Assistants"}, []string{"Web App", "Mobile App", "API"},
		},
		{
			"Gemini", "Google's multimodal AI assistant",
			"Gemini is Google's flagship AI assistant, deeply integrated with Google Workspace and Search.\n\n## Key features\n\n- **Gemini 2.0** with native multimodal input (text, images, audio, video)\n- **Deep Research** for multi-step research tasks with cited sources\n- **Google Workspace integration** — summarize Gmail, Docs, Drive\n- Available on web, iOS, and Android\n\nGemini shines for users already in the Google ecosystem, combining the power of Google Search with conversational AI.",
			"https://gemini.google.com", true,
			[]string{"Chatbots & Assistants"}, []string{"Web App", "Mobile App", "API"},
		},
		{
			"Mistral Chat", "Open-source conversational AI",
			"Mistral Chat provides access to Mistral AI's family of open-weight models, known for efficiency and strong reasoning.\n\n## Key features\n\n- **Open-weight models** — run locally or via API\n- **Mistral Large** for complex reasoning and coding\n- **Le Chat** web interface with document upload\n- API access compatible with OpenAI client libraries\n\nMistral's models are a top choice for developers who need powerful AI without vendor lock-in.",
			"https://chat.mistral.ai", true,
			[]string{"Chatbots & Assistants"}, []string{"Web App", "API", "AI Model", "Open Source"},
		},
		{
			"Pi", "Personal AI companion by Inflection",
			"Pi is a personal AI companion from Inflection AI, designed for warm, supportive conversations.\n\n## Key features\n\n- **Empathetic conversation style** — great for thinking out loud\n- **Remembers context** across conversations\n- **Voice mode** on mobile for natural spoken dialogue\n- Available on web, iOS, and Android\n\nPi is best for users looking for a thoughtful, supportive conversational partner rather than a task-oriented assistant.",
			"https://pi.ai", true,
			[]string{"Chatbots & Assistants"}, []string{"Web App", "Mobile App"},
		},
		// Image Generation (6)
		{
			"Midjourney", "High-quality AI art from text prompts",
			"Midjourney produces some of the most visually stunning AI-generated images available, favored by artists and designers.\n\n## Key features\n\n- **v6.1 model** with photorealistic and stylized output\n- **Style references** and **character consistency** across generations\n- **Vary (Region)** for inpainting and localized edits\n- Web interface with image gallery and prompt history\n\nMidjourney is the go-to choice for high-quality AI art, particularly for editorial, concept art, and creative projects.",
			"https://midjourney.com", false,
			[]string{"Image Generation", "Art Generation"}, []string{"Web App"},
		},
		{
			"DALL-E 3", "OpenAI's latest image generation model",
			"DALL-E 3 is OpenAI's image generation model, integrated into ChatGPT and available via API.\n\n## Key features\n\n- **Precise prompt following** — understands complex, detailed descriptions\n- **Text rendering** in images (one of the best in class)\n- **Integrated with ChatGPT** — use natural conversation to generate and iterate\n- Available via OpenAI API with `gpt-image-1`\n\nDALL-E 3 is ideal for generating images from detailed natural language prompts, especially when text accuracy in images matters.",
			"https://openai.com/dall-e-3", false,
			[]string{"Image Generation"}, []string{"API", "AI Model"},
		},
		{
			"Stable Diffusion", "Open-source AI image generation",
			"Stable Diffusion is the leading open-source image generation model, powering thousands of apps and custom deployments.\n\n## Key features\n\n- **Fully open weights** — run locally on consumer hardware\n- **Huge ecosystem** of fine-tuned models, LoRAs, and extensions\n- **ControlNet** for pose, depth, and edge-guided generation\n- Supports text-to-image, image-to-image, and inpainting\n\nStable Diffusion gives developers and power users maximum control and flexibility, with no usage costs when self-hosted.",
			"https://stability.ai", true,
			[]string{"Image Generation", "Art Generation"}, []string{"AI Model", "Open Source", "Python Library"},
		},
		{
			"Adobe Firefly", "Creative AI built into Adobe's suite",
			"Adobe Firefly is Adobe's family of generative AI models, built for commercial safety and deep Creative Cloud integration.\n\n## Key features\n\n- **Commercially safe** — trained on licensed Adobe Stock images\n- **Generative Fill** in Photoshop for seamless inpainting\n- **Text Effects** and **Generative Recolor** in Illustrator\n- **Firefly Vector** for scalable AI-generated vector art\n\nFirefly is the best choice for creative professionals who need AI generation that's safe for commercial use and works inside their existing Adobe workflow.",
			"https://firefly.adobe.com", true,
			[]string{"Image Generation", "Photo Editing"}, []string{"Web App", "Desktop App"},
		},
		{
			"Leonardo.ai", "AI image generation for game assets and art",
			"Leonardo.ai is an AI image generation platform optimized for game development, concept art, and consistent character creation.\n\n## Key features\n\n- **Phoenix** and **Flux** models with fine control over style\n- **AI Canvas** for inpainting and outpainting\n- **Motion** for animating still images\n- **Custom model training** on your own assets\n\nLeonardo.ai is especially popular with game developers and digital artists who need style-consistent, production-ready assets.",
			"https://leonardo.ai", true,
			[]string{"Image Generation", "Art Generation"}, []string{"Web App", "API"},
		},
		{
			"Ideogram", "Text-accurate AI image generation",
			"Ideogram is an AI image generator purpose-built for accurately rendering text inside images.\n\n## Key features\n\n- **Best-in-class text rendering** inside generated images\n- **Magic Prompt** for automatic prompt enhancement\n- **Remix** to iterate on existing images\n- **Style presets** including photo, illustration, and typography\n\nIdeogram is the top choice when your generated image needs to include readable, accurate text — logos, posters, social media graphics.",
			"https://ideogram.ai", true,
			[]string{"Image Generation"}, []string{"Web App", "API"},
		},
		// Code & Development (6)
		{
			"GitHub Copilot", "AI pair programmer in your editor",
			"GitHub Copilot is the most widely used AI coding assistant, integrated directly into VS Code, JetBrains IDEs, and more.\n\n## Key features\n\n- **Autocomplete** with multi-line suggestions as you type\n- **Copilot Chat** for inline Q&A and code explanations\n- **Copilot Workspace** for agent-based task completion\n- **Pull Request summaries** and code review suggestions\n\nCopilot is the industry standard for AI-assisted development, with the largest ecosystem of IDE integrations and the deepest GitHub integration.",
			"https://github.com/features/copilot", true,
			[]string{"Code & Development", "Code Completion"}, []string{"Browser Extension", "API"},
		},
		{
			"Cursor", "AI-first code editor",
			"Cursor is a VS Code fork rebuilt from the ground up for AI-assisted development, with deep codebase understanding.\n\n## Key features\n\n- **Composer** for multi-file edits from a single prompt\n- **Codebase indexing** — ask questions about your entire project\n- **Tab completion** that predicts your next edit, not just the next line\n- **Agent mode** for autonomous, multi-step coding tasks\n\nCursor has become the preferred editor for developers who want the most capable AI coding experience available.",
			"https://cursor.com", true,
			[]string{"Code & Development", "Code Completion"}, []string{"Desktop App"},
		},
		{
			"Tabnine", "AI code completion for any IDE",
			"Tabnine is an AI code completion tool that works across 30+ programming languages and all major IDEs.\n\n## Key features\n\n- **Privacy-first** option to run models locally or on-prem\n- **Personalized models** trained on your team's codebase\n- **Code review agent** for automated PR feedback\n- Supports VS Code, JetBrains, Neovim, Eclipse, and more\n\nTabnine is favored by enterprise teams that need powerful AI completion with strict data privacy and on-premises deployment options.",
			"https://tabnine.com", true,
			[]string{"Code & Development", "Code Completion"}, []string{"Desktop App", "Browser Extension"},
		},
		{
			"Codeium", "Free AI code completion and chat",
			"Codeium offers fast, free AI code completion and chat for individual developers, with generous limits.\n\n## Key features\n\n- **Free forever** for individual developers\n- **Windsurf IDE** — Codeium's own AI-first editor\n- **Context-aware completions** that understand your whole project\n- Supports 70+ languages and all major editors\n\nCodeium is the best free alternative to GitHub Copilot, with no token limits and a growing set of agentic features.",
			"https://codeium.com", true,
			[]string{"Code & Development", "Code Completion"}, []string{"Web App", "Desktop App", "Browser Extension"},
		},
		{
			"Phind", "AI search engine for developers",
			"Phind is an AI-powered search engine built specifically for programmers, combining web search with code-aware LLMs.\n\n## Key features\n\n- **Searches the web** and synthesizes developer-focused answers\n- **VS Code extension** for in-editor AI search\n- **Phind-70B** model fine-tuned on code and technical content\n- Pair programmer mode for step-by-step problem solving\n\nPhind is the best search tool for developers who want answers that go beyond documentation — with real code examples from the latest sources.",
			"https://phind.com", true,
			[]string{"Code & Development", "Search & Research"}, []string{"Web App"},
		},
		{
			"Replit AI", "AI-powered collaborative coding environment",
			"Replit is a browser-based IDE with deep AI integration, enabling anyone to build and deploy software without setup.\n\n## Key features\n\n- **Replit Agent** — describe an app in plain English and it builds it\n- **Instant deployment** — share a live URL the moment you start coding\n- **Multiplayer** for real-time collaborative coding\n- Supports 50+ languages with zero environment setup\n\nReplit is the fastest way to go from idea to deployed app, especially for learners, prototypers, and developers who don't want to deal with infrastructure.",
			"https://replit.com", true,
			[]string{"Code & Development"}, []string{"Web App", "Mobile App"},
		},
		// Writing & Content (5)
		{
			"Jasper", "AI writing platform for marketing teams",
			"Jasper is an enterprise AI writing platform built for marketing teams, with brand voice controls and workflow integrations.\n\n## Key features\n\n- **Brand Voice** — train Jasper on your brand's tone and style\n- **Campaigns** for generating a full suite of content from one brief\n- **Jasper Art** for AI image generation alongside written content\n- Integrates with Surfer SEO, Webflow, and HubSpot\n\nJasper is the leading choice for marketing teams that need brand-consistent content at scale, with the collaboration and review workflows enterprises expect.",
			"https://jasper.ai", false,
			[]string{"Writing & Content", "Copywriting"}, []string{"Web App", "Browser Extension"},
		},
		{
			"Copy.ai", "AI copywriting and content assistant",
			"Copy.ai is an AI content platform focused on go-to-market teams, from sales emails to blog posts.\n\n## Key features\n\n- **GTM AI Platform** with automated workflows for sales and marketing\n- **80+ templates** covering every content format\n- **Brand voice** settings for consistent tone\n- Free plan available with no credit card required\n\nCopy.ai is popular with startups and solo marketers who need versatile AI writing across many formats without a large budget.",
			"https://copy.ai", true,
			[]string{"Writing & Content", "Copywriting"}, []string{"Web App"},
		},
		{
			"Writesonic", "AI content creation for blogs and ads",
			"Writesonic is an AI writing platform that covers everything from long-form blog posts to paid ad copy.\n\n## Key features\n\n- **Chatsonic** — a web-connected AI chat with image generation\n- **Article Writer 6** for SEO-optimized long-form content\n- **Botsonic** for building AI chatbots from your own data\n- Integrates with Semrush, Zapier, and WordPress\n\nWritesonic is a good all-in-one option for content marketers who want one tool for articles, ads, chatbots, and social media.",
			"https://writesonic.com", true,
			[]string{"Writing & Content", "Copywriting"}, []string{"Web App", "API"},
		},
		{
			"Grammarly", "AI writing assistant for clarity and correctness",
			"Grammarly is the leading AI writing assistant, helping over 30 million people write with greater clarity and confidence.\n\n## Key features\n\n- **Real-time grammar, spelling, and style suggestions** everywhere you write\n- **Generative AI** for drafting, rewriting, and summarizing text\n- **Tone detection** and adjustment for professional or casual contexts\n- Works across 500,000+ apps and websites via browser extension\n\nGrammarly is the most ubiquitous writing AI, trusted from students to Fortune 500 companies for its reliable, always-on assistance.",
			"https://grammarly.com", true,
			[]string{"Writing & Content"}, []string{"Web App", "Desktop App", "Browser Extension"},
		},
		{
			"DeepL", "AI-powered translation with nuance",
			"DeepL delivers the highest quality machine translation available, with a level of nuance that sets it apart from alternatives.\n\n## Key features\n\n- **DeepL Translator** supporting 30+ languages with exceptional accuracy\n- **DeepL Write** for AI-powered grammar and style improvement\n- **DeepL API** for integrating translation into any application\n- **Glossary** feature for consistent terminology in professional documents\n\nDeepL is the preferred translation tool for professionals, legal teams, and publishers where linguistic accuracy and natural-sounding output are non-negotiable.",
			"https://deepl.com", true,
			[]string{"Writing & Content", "Translation"}, []string{"Web App", "Desktop App", "API"},
		},
		// Video & Animation (5)
		{
			"Runway", "AI video creation and editing suite",
			"Runway is a creative AI platform leading the text-to-video space, used by filmmakers, studios, and content creators.\n\n## Key features\n\n- **Gen-3 Alpha** for high-fidelity text-to-video and image-to-video\n- **Motion Brush** for directing movement in generated clips\n- **Inpainting and masking** for precise video editing\n- **Act One** for facial performance capture\n\nRunway is the professional's choice for AI video generation, with the most advanced motion control and editing features available.",
			"https://runwayml.com", true,
			[]string{"Video & Animation", "Text-to-Video"}, []string{"Web App", "API"},
		},
		{
			"Pika", "Generate and edit video with AI",
			"Pika is an accessible AI video platform that makes it easy for anyone to generate and modify video clips.\n\n## Key features\n\n- **Pika 2.2** with improved motion quality and scene consistency\n- **Pikaffects** for creative transformations like explosions and deflation\n- **Lip sync** for animating characters with audio\n- Available on web and iOS\n\nPika is popular with social media creators and marketers who want to quickly produce eye-catching short-form AI video without a steep learning curve.",
			"https://pika.art", true,
			[]string{"Video & Animation", "Text-to-Video"}, []string{"Web App", "Mobile App"},
		},
		{
			"HeyGen", "AI video avatar and dubbing platform",
			"HeyGen enables businesses to create professional presenter videos with AI avatars and lifelike dubbing.\n\n## Key features\n\n- **300+ AI avatars** or create a custom avatar from your own video\n- **Video Translation** — dub videos into 40+ languages with lip sync\n- **Interactive Avatar** API for real-time AI video conversations\n- **Streaming Avatar** for live video calls powered by AI\n\nHeyGen is the leader in AI avatar video, widely used for corporate training, marketing videos, and multilingual content localization.",
			"https://heygen.com", true,
			[]string{"Video & Animation"}, []string{"Web App", "API"},
		},
		{
			"Synthesia", "Create AI presenter videos at scale",
			"Synthesia is an enterprise AI video platform for creating professional presenter videos without cameras or studios.\n\n## Key features\n\n- **230+ AI avatars** across diverse backgrounds and styles\n- **Custom avatar** creation from a short video clip\n- **120+ languages** with natural-sounding AI voiceovers\n- **SCORM export** for learning management system integration\n\nSynthesia is widely adopted in corporate L&D, HR, and internal communications for producing consistent, professional video content at scale.",
			"https://synthesia.io", false,
			[]string{"Video & Animation"}, []string{"Web App", "API"},
		},
		{
			"D-ID", "Animate photos with AI-generated speech",
			"D-ID brings static images to life by combining them with AI-generated speech and realistic facial animation.\n\n## Key features\n\n- **Photo animation** — make any face speak with natural lip sync\n- **Text-to-video** from a script and an image in seconds\n- **Creative Reality Studio** for rapid talking-head video production\n- **Streaming API** for real-time interactive AI avatars\n\nD-ID is popular for social media content, digital twins, and building interactive AI agents with a face.",
			"https://d-id.com", true,
			[]string{"Video & Animation"}, []string{"Web App", "API"},
		},
		// Audio & Music (6)
		{
			"ElevenLabs", "Realistic AI voice synthesis and cloning",
			"ElevenLabs produces the most realistic AI voice synthesis available, with instant voice cloning from a short audio sample.\n\n## Key features\n\n- **Voice cloning** from as little as one minute of audio\n- **29 languages** with natural prosody and emotion\n- **AI Dubbing** for translating video while preserving the original voice\n- **Conversational AI** API for building voice-enabled agents\n\nElevenLabs is the gold standard for AI voice, used by publishers, game developers, filmmakers, and developers building voice products.",
			"https://elevenlabs.io", true,
			[]string{"Audio & Music", "Voice & Speech"}, []string{"Web App", "Mobile App", "API"},
		},
		{
			"Suno", "Generate full songs with AI",
			"Suno generates complete, production-quality songs — vocals, instruments, and mixing — from a simple text prompt.\n\n## Key features\n\n- **Full song generation** with lyrics, vocals, and instrumentation\n- **Covers any genre** from pop to opera to death metal\n- **Extend and remix** existing tracks\n- Available on web and mobile with a free daily credit allowance\n\nSuno has made AI music creation accessible to everyone, producing songs that sound surprisingly polished for casual and creative use.",
			"https://suno.com", true,
			[]string{"Audio & Music"}, []string{"Web App", "Mobile App"},
		},
		{
			"Udio", "AI music creation from text descriptions",
			"Udio creates high-fidelity AI music with a focus on audio quality and detailed genre control.\n\n## Key features\n\n- **High-fidelity output** — often indistinguishable from real recordings\n- **Precise genre and mood tags** for fine-tuned control\n- **Extend tracks** by generating additional sections\n- Custom lyrics input or AI-generated\n\nUdio is preferred by music enthusiasts who prioritize audio quality and want granular control over the style and feel of generated tracks.",
			"https://udio.com", true,
			[]string{"Audio & Music"}, []string{"Web App"},
		},
		{
			"Whisper", "OpenAI's open-source speech recognition",
			"Whisper is OpenAI's open-source automatic speech recognition model, accurate across accents, languages, and noisy environments.\n\n## Key features\n\n- **99 languages** with state-of-the-art transcription accuracy\n- **Timestamps** at the word and segment level\n- **Translation** from any supported language to English\n- Runs locally — no API costs or data privacy concerns\n\nWhisper is the most widely deployed open-source transcription model, used in applications from meeting tools to subtitle generators to voice interfaces.",
			"https://openai.com/research/whisper", true,
			[]string{"Audio & Music", "Voice & Speech"}, []string{"AI Model", "Open Source", "Python Library", "CLI Tool"},
		},
		{
			"Mubert", "AI-generated royalty-free music streams",
			"Mubert generates royalty-free background music in real time, tailored to mood, activity, or content type.\n\n## Key features\n\n- **Text-to-music** generation for any mood or genre\n- **Royalty-free licensing** for content creators and developers\n- **API** for generating music programmatically\n- Tracks are unique each time — no repeated loops\n\nMubert is popular with video creators, streamers, and app developers who need background music that's always fresh and legally safe to use.",
			"https://mubert.com", true,
			[]string{"Audio & Music"}, []string{"Web App", "API"},
		},
		{
			"Descript", "AI-powered audio and video editing",
			"Descript lets you edit audio and video by editing a text transcript — a fundamentally different approach to media editing.\n\n## Key features\n\n- **Text-based editing** — cut words from the transcript to cut from the video\n- **Overdub** for cloning your own voice to fix mistakes\n- **Studio Sound** for one-click audio cleanup and noise removal\n- **Screen recording** and podcast publishing built in\n\nDescript is the fastest editor for podcasters, YouTubers, and course creators who want to produce polished content without learning traditional editing software.",
			"https://descript.com", true,
			[]string{"Audio & Music", "Video & Animation"}, []string{"Desktop App", "Web App"},
		},
		// Data & Analytics (5)
		{
			"Obviously AI", "No-code AI predictions for business data",
			"Obviously AI lets non-technical users build predictive models from CSV or database data with no coding required.\n\n## Key features\n\n- **One-click prediction models** from any structured dataset\n- **Natural language queries** — ask questions about your data\n- **Automated feature engineering** and model selection\n- Integrations with Google Sheets, Salesforce, and more\n\nObviously AI is built for business analysts and growth teams who need predictive AI without data science expertise.",
			"https://obviously.ai", false,
			[]string{"Data & Analytics", "Business Intelligence"}, []string{"Web App"},
		},
		{
			"Akkio", "AI analytics platform for growth teams",
			"Akkio is an AI-powered analytics platform that helps agencies and business teams turn data into decisions faster.\n\n## Key features\n\n- **Chat Explore** for conversational data analysis\n- **Predictive modeling** without code\n- **AI-generated reports** for client presentations\n- Connects to HubSpot, Salesforce, Google Analytics, and more\n\nAkkio is popular with marketing agencies and ops teams that need fast data insights without a dedicated data science team.",
			"https://akkio.com", false,
			[]string{"Data & Analytics", "Business Intelligence"}, []string{"Web App"},
		},
		{
			"Airtable AI", "Database and workflow automation with AI",
			"Airtable AI brings generative AI directly into Airtable's flexible database and workflow platform.\n\n## Key features\n\n- **AI fields** that summarize, categorize, or extract data automatically\n- **Automations** triggered by AI-evaluated conditions\n- **Connected tables** for building relational workflows\n- Works alongside existing Airtable views, forms, and integrations\n\nAirtable AI is ideal for operations teams that already use Airtable and want to add intelligent automation without leaving their existing workspace.",
			"https://airtable.com", true,
			[]string{"Data & Analytics", "Productivity"}, []string{"Web App", "Mobile App"},
		},
		{
			"Julius", "Chat with your data using AI",
			"Julius is an AI data analyst that connects to your data sources and lets you explore them through natural language conversation.\n\n## Key features\n\n- **CSV, Excel, and Google Sheets** upload with instant analysis\n- **Automatic chart generation** from conversational queries\n- **Statistical analysis** without writing code\n- **Python and R code export** for reproducible analysis\n\nJulius is the fastest way for non-engineers to do serious data analysis — just describe what you want to know and it does the work.",
			"https://julius.ai", true,
			[]string{"Data & Analytics"}, []string{"Web App"},
		},
		{
			"MonkeyLearn", "No-code text analysis and classification",
			"MonkeyLearn is a no-code text analysis platform for building custom classifiers, extractors, and sentiment models.\n\n## Key features\n\n- **Custom text classifiers** trained on your own labeled data\n- **Sentiment analysis**, keyword extraction, and topic classification\n- **Pre-built models** ready to use immediately\n- Integrates with Zapier, Google Sheets, and via API\n\nMonkeyLearn is used by customer success teams and researchers to automatically categorize support tickets, survey responses, and user reviews at scale.",
			"https://monkeylearn.com", false,
			[]string{"Data & Analytics"}, []string{"Web App", "API"},
		},
		// Productivity (5)
		{
			"Notion AI", "AI writing and summarization inside Notion",
			"Notion AI is built directly into Notion's workspace, helping teams draft, summarize, and act on information without switching apps.\n\n## Key features\n\n- **AI writing** for drafting pages, documents, and meeting notes\n- **Summarize** long pages or meeting notes instantly\n- **Q&A** — ask questions and get answers sourced from your workspace\n- **Autofill database properties** with AI-generated values\n\nNotion AI is the best choice for teams already on Notion who want AI that understands the context of their entire knowledge base.",
			"https://notion.so", true,
			[]string{"Productivity", "Note Taking"}, []string{"Web App", "Desktop App", "Mobile App"},
		},
		{
			"Mem", "AI-powered self-organizing notes",
			"Mem is an AI note-taking app that automatically surfaces relevant information from your past notes when you need it.\n\n## Key features\n\n- **AI-powered search** across all your notes and linked content\n- **Smart collections** that organize notes automatically by topic\n- **Write with AI** for drafting content inside your note history\n- Captures from email, Slack, and browser extensions\n\nMem is built for knowledge workers who take a lot of notes and want AI to connect the dots rather than filing things manually.",
			"https://mem.ai", false,
			[]string{"Productivity", "Note Taking"}, []string{"Web App", "Mobile App"},
		},
		{
			"Otter.ai", "AI meeting transcription and summaries",
			"Otter.ai transcribes meetings in real time and generates summaries, action items, and follow-up emails automatically.\n\n## Key features\n\n- **Real-time transcription** for Zoom, Google Meet, and Teams\n- **AI meeting summaries** with action items extracted automatically\n- **OtterPilot** joins meetings on your behalf and takes notes\n- **Speaker identification** for accurate attribution\n\nOtter.ai is the most widely used AI meeting tool, trusted by teams who spend a lot of time in video calls and need accurate, searchable records.",
			"https://otter.ai", true,
			[]string{"Productivity", "Voice & Speech"}, []string{"Web App", "Mobile App"},
		},
		{
			"Motion", "AI calendar, task, and project manager",
			"Motion uses AI to automatically schedule your tasks and meetings, building the optimal daily plan based on priorities and deadlines.\n\n## Key features\n\n- **Automatic scheduling** — Motion builds your day for you\n- **Dynamic rescheduling** when priorities shift or meetings run long\n- **Project management** with AI-managed timelines\n- **Meeting booking links** that protect your focus time\n\nMotion is for high-output professionals and teams who are overwhelmed by task management and want AI to handle the scheduling decisions for them.",
			"https://usemotion.com", false,
			[]string{"Productivity"}, []string{"Web App", "Mobile App"},
		},
		{
			"Reclaim.ai", "AI scheduling assistant for busy teams",
			"Reclaim.ai protects your habits, focus time, and priorities by automatically scheduling them into your Google Calendar.\n\n## Key features\n\n- **Habit scheduling** — Reclaim finds the best time for recurring blocks\n- **Task syncing** with Asana, Todoist, Linear, and more\n- **Smart 1:1s** that flexibly schedule regular check-ins\n- **Scheduling links** that respect your protected time\n\nReclaim.ai is popular with remote workers and engineering teams who use Google Calendar and want AI to defend their deep work time.",
			"https://reclaim.ai", true,
			[]string{"Productivity"}, []string{"Web App", "Browser Extension"},
		},
		// Search & Research (2)
		{
			"Perplexity", "AI-powered search with cited answers",
			"Perplexity reimagines search as a conversational experience, providing direct answers with cited sources from the live web.\n\n## Key features\n\n- **Real-time web search** with transparent citations\n- **Deep Research** for in-depth, multi-source research reports\n- **Spaces** for collaborative, topic-focused research\n- **Pro Search** with access to GPT-4o, Claude, and Sonar Large\n\nPerplexity is the fastest-growing AI search tool, favored by researchers, journalists, and anyone who wants answers that are both fast and verifiable.",
			"https://perplexity.ai", true,
			[]string{"Search & Research"}, []string{"Web App", "Mobile App", "API"},
		},
		{
			"You.com", "AI search engine and assistant",
			"You.com is an AI-powered search engine that combines web results with AI summarization and a customizable interface.\n\n## Key features\n\n- **AI summarization** layered on top of traditional web search\n- **Research Mode** for multi-step, cited research reports\n- **Code Mode** for programming questions with runnable snippets\n- **Custom apps** to choose your preferred AI models and tools\n\nYou.com is for users who want more control over their AI search experience, with the flexibility to choose different AI models for different tasks.",
			"https://you.com", true,
			[]string{"Search & Research", "Chatbots & Assistants"}, []string{"Web App"},
		},
		// Design & UX (5)
		{
			"Framer AI", "AI-powered website and landing page builder",
			"Framer AI generates fully responsive, interactive websites from a text prompt, built on Framer's professional design platform.\n\n## Key features\n\n- **AI site generation** from a single text description\n- **Full Framer editor** for pixel-perfect customization after generation\n- **CMS and localization** for dynamic, multi-language content\n- **One-click publishing** with built-in hosting on framer.app\n\nFramer AI is the fastest path from an idea to a live, professional-looking website — without touching code unless you want to.",
			"https://framer.com", true,
			[]string{"Design & UX"}, []string{"Web App"},
		},
		{
			"Uizard", "Turn sketches into UI designs with AI",
			"Uizard turns rough wireframes, screenshots, or text descriptions into editable UI designs in seconds.\n\n## Key features\n\n- **Screenshot Scanner** — upload a UI screenshot and get an editable version\n- **Text to UI** — describe a screen and get a wireframe or mockup\n- **Autodesigner** for generating complete multi-screen app designs\n- **Real-time collaboration** for design teams\n\nUizard is the fastest way to prototype an app or website, especially for product managers and developers who don't have a design background.",
			"https://uizard.io", true,
			[]string{"Design & UX"}, []string{"Web App"},
		},
		{
			"v0", "Generate UI components with AI by Vercel",
			"v0 generates copy-paste React and Tailwind UI components from a text prompt, built by Vercel.\n\n## Key features\n\n- **React + Tailwind** components ready to paste into any project\n- **Shadcn/ui** based components for a consistent design system\n- **Iterate with chat** — refine components through conversation\n- **Publish and share** components with a public URL\n\nv0 is the go-to tool for frontend developers who want to skip the boilerplate and get a polished UI component in seconds.",
			"https://v0.dev", true,
			[]string{"Design & UX", "Code & Development"}, []string{"Web App", "API"},
		},
		{
			"Canva AI", "AI design features inside Canva",
			"Canva AI brings a suite of generative AI features into Canva's visual design platform, used by over 170 million people.\n\n## Key features\n\n- **Magic Studio** with text-to-image, background remover, and object eraser\n- **Magic Write** for AI copywriting inside designs\n- **Magic Animate** for one-click presentation animations\n- **Brand Kit** for consistent colors, fonts, and logos across AI generations\n\nCanva AI makes advanced design and AI generation accessible to everyone, embedded in the tool that non-designers already use every day.",
			"https://canva.com", true,
			[]string{"Design & UX", "Image Generation"}, []string{"Web App", "Mobile App", "Desktop App"},
		},
		{
			"Galileo AI", "Generate editable UI designs from text",
			"Galileo AI generates complex, multi-component UI designs from a text description, ready to import into Figma.\n\n## Key features\n\n- **Full-fidelity UI generation** from detailed text prompts\n- **Figma export** for editing in your existing design workflow\n- **Component-aware** output that follows design system conventions\n- Generates screens, not just individual components\n\nGalileo AI is aimed at product designers who want to rapidly explore design directions without spending hours on initial mockups.",
			"https://usegalileo.ai", false,
			[]string{"Design & UX"}, []string{"Web App"},
		},
	}

	aiObjs := make([]*models.AI, 0, len(aiDefs))
	for _, def := range aiDefs {
		slug := toSlug(def.Name)
		var a models.AI
		db.Where("slug = ?", slug).First(&a)
		a.Name = def.Name
		a.Slug = slug
		a.Subtitle = def.Subtitle
		a.Description = def.Description
		a.Website = def.Website
		a.HaveFree = def.HaveFree
		if err := upsert(db, a.ID, &a); err != nil {
			log.Fatalf("upsert ai %q: %v", def.Name, err)
		}
		cats := make([]models.Category, 0, len(def.Categories))
		for _, cn := range def.Categories {
			if c, ok := catMap[cn]; ok {
				cats = append(cats, *c)
			}
		}
		if err := db.Model(&a).Association("Categories").Replace(cats); err != nil {
			log.Fatalf("assign categories for %q: %v", def.Name, err)
		}
		genera := make([]models.Genus, 0, len(def.Genera))
		for _, gn := range def.Genera {
			if g, ok := genusMap[gn]; ok {
				genera = append(genera, *g)
			}
		}
		if err := db.Model(&a).Association("Genera").Replace(genera); err != nil {
			log.Fatalf("assign genera for %q: %v", def.Name, err)
		}
		aiObjs = append(aiObjs, &a)
	}
	fmt.Printf("  ais: %d\n", len(aiDefs))

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
		Slug     string
		Subtitle string
		AIIdxs   []int
	}

	listDefs := []listDef{
		{"Best AI Chatbots", "best-ai-chatbots", "The top conversational AI assistants you can use today", []int{0, 1, 2, 3, 4}},
		{"Top Image Generators", "top-image-generators", "AIs that turn text into stunning visuals", []int{5, 6, 7, 8, 9, 10}},
		{"AI Coding Assistants", "ai-coding-assistants", "Write, complete, and review code faster with AI", []int{11, 12, 13, 14, 15, 16}},
		{"AI Writing Tools", "ai-writing-tools", "Write better content with AI assistance", []int{17, 18, 19, 20, 21}},
		{"AI Video Generators", "ai-video-generators", "Create and edit video entirely with AI", []int{22, 23, 24, 25, 26}},
		{"AI Audio & Music", "ai-audio-music", "Voice, music, and sound AIs", []int{27, 28, 29, 30, 31, 32}},
		{"AI Productivity Apps", "ai-productivity-apps", "Work smarter with AI-powered productivity AIs", []int{38, 39, 40, 41, 42}},
		{"AI Data Tools", "ai-data-tools", "Analyze and visualize data using AI", []int{33, 34, 35, 36, 37}},
		{"Free AI Tools", "free-ai-tools", "Powerful AIs you can start using for free", []int{1, 4, 7, 14, 30, 39, 43}},
		{"Getting Started with AI", "getting-started-with-ai", "The best AIs for beginners", []int{0, 1, 5, 11, 17, 38, 43}},
	}

	for _, def := range listDefs {
		var l models.List
		db.Where("slug = ?", def.Slug).First(&l)
		l.Name = def.Name
		l.Slug = def.Slug
		l.Subtitle = def.Subtitle
		if err := upsert(db, l.ID, &l); err != nil {
			log.Fatalf("upsert list %q: %v", def.Name, err)
		}
		db.Where("list_id = ?", l.ID).Delete(&models.ListAI{})
		for sort, ti := range def.AIIdxs {
			lt := models.ListAI{ListID: l.ID, AIID: aiObjs[ti].ID, Sort: sort}
			if err := db.Create(&lt).Error; err != nil {
				log.Fatalf("create list_ai: %v", err)
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
