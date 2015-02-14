package sosoinit

import(
	"log"
	"os"
	"../context"
	"../configure"
	"../scheduler"
	"../downloader"
	"../analyzer"
	"../utils/db"
)

const(
	SCHEDULER = "scheduler"
	DOWNLOADER = "downloader"
	ANALYZER = "analyzer"
	DATABASE = "database"

	MASTER = "master"
	PORT = "port"
	DOWNLOAD_PATH = "download_path"
	DICTIONARY_PATH = "dictionary_path"
	STOPWORDS_PATH = "stopwords_path"
)

func Sosoinit(context *context.Context) {
	var scheduler *scheduler.Scheduler
	var downloaders []*downloader.Downloader
	var analyzers []*analyzer.Analyzer
	config := configure.InitConfig("./config.ini")

	if schedulerConfig,exist := config.GetEntity(SCHEDULER);exist {
		checkSchedulerConfig(schedulerConfig)
		scheduler = initScheduler(context, schedulerConfig)
	}

	if downloaderConfig,exist := config.GetEntity(DOWNLOADER);exist {
		checkDownloaderConfig(downloaderConfig)
		downloaders = initDownloaders(context, downloaderConfig)
	}

	if analyzerConfig,exist := config.GetEntity(ANALYZER);exist {
		checkAnalyzerConfig(analyzerConfig)
		analyzers = initAnalyzers(context, analyzerConfig)
	}

	if dbConfig,exist := config.GetEntity(DATABASE);exist {
		checkDatabaseConfig(dbConfig)
	} else {
		log.Fatal("缺少数据库配置")
	}

	context.AddComponent("scheduler", scheduler)
	context.AddComponent("downloaders", downloaders)
	context.AddComponent("analyzers", analyzers)
}

/**
 * 检查调度器的配置是否正确
 */
func checkSchedulerConfig(es []*configure.Entity) {
	if len(es) > 1 {
		//调度器只能有一个
		log.Fatal("scheduler配置重复")
	}
}

/**
 * 检查下载器的配置是否正确
 */
func checkDownloaderConfig(es []*configure.Entity) {
	for i,e := range es {
		if e.GetAttr(MASTER) == "" {
			log.Fatal("存在下载器没有对应master")
		}

		if e.GetAttr(DOWNLOAD_PATH) != "" {
			dir := e.GetAttr(DOWNLOAD_PATH)

			fi, err := os.Stat(dir)
			if err != nil && !os.IsExist(err) || !fi.IsDir() {
				err := os.MkdirAll(dir,0777)
				if err != nil {
					log.Fatalf("第%d个下载器的下载路径无法生成\n", i)
				}
			}
		} else {
			log.Fatalf("第%d个下载器没有配置下载路径\n", i)
		}
	}
}

/**
 * 检查分析器的配置是否正确
 */
func checkAnalyzerConfig(es []*configure.Entity) {
	for i,e := range es {
		if e.GetAttr(MASTER) == "" {
			log.Fatal("存在分析器没有对应master")
		}
		if e.GetAttr(DICTIONARY_PATH) != "" {
			dictionary := e.GetAttr(DICTIONARY_PATH)
			fi, err := os.Stat(dictionary)
			if err != nil && !os.IsExist(err) || fi.IsDir() {
				log.Fatalf("第%d个分析器的词典路径错误\n", i)
			}
		} else {
			log.Fatalf("第%d个分析器没有配置词典路径\n", i)
		}

		if e.GetAttr(STOPWORDS_PATH) != "" {
			stopwords := e.GetAttr(STOPWORDS_PATH)
			fi, err := os.Stat(stopwords)
			if err != nil && !os.IsExist(err) || fi.IsDir() {
				log.Fatalf("第%d个分析器的停用词典路径错误\n", i)
			}
		} else {
			log.Fatalf("第%d个分析器没有配置停用词典路径\n", i)
		}
	}
}

/**
 * 检查数据库配置是否正确
 */
func checkDatabaseConfig(es []*configure.Entity) {
	if len(es) > 1 {
		log.Fatal("数据库配置重复")
	}
	e := es[0]
	host := e.GetAttr("host")
	username := e.GetAttr("username")
	password := e.GetAttr("password")
	dbname := e.GetAttr("dbname")
	charset := e.GetAttr("charset")

	m := make(map[string]string)
	m["host"] = host
	m["username"] = username
	m["password"] = password
	m["dbname"] = dbname
	m["charset"] = charset

	databaseConfig := db.New(m)
	if exist,_ := databaseConfig.CheckDBExist();!exist {
		log.Fatal("数据库链接失败")
	}
}

/**
 * 初始化调度器
 */
func initScheduler(context *context.Context, es []*configure.Entity) *scheduler.Scheduler{
	e := es[0]
	scheduler := scheduler.New(context, e.GetAttr(PORT))
	return scheduler
}

func initDownloaders(context *context.Context, es []*configure.Entity) []*downloader.Downloader {
	downloaders := make([]*downloader.Downloader, 0)
	for _,e := range es {
		d := downloader.New(context, e.GetAttr(MASTER), e.GetAttr(DOWNLOAD_PATH))
		downloaders = append(downloaders, d)
	}
	return downloaders
}

func initAnalyzers(context *context.Context, es []*configure.Entity) []*analyzer.Analyzer {
	analyzers := make([]*analyzer.Analyzer, 0)
	for _,e := range es {
		a := analyzer.New(context, e.GetAttr(MASTER), e.GetAttr(DICTIONARY_PATH), e.GetAttr(STOPWORDS_PATH))
		analyzers = append(analyzers, a)
	}
	return analyzers
}