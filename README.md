# Go-MySQL

How to connect your golang app with this driver

![Go-MySQL-Driver logo](https://raw.github.com/wiki/go-sql-driver/mysql/gomysql_m.png "Golang Gopher holding the MySQL Dolphin")

func SetupDatabase(waitGroup *sync.WaitGroup){
	//Setting up geo database
	var geoErr error
	geoDb, geoErr = geoip2.Open("GeoLite2-City.mmdb")
	if geoErr != nil {

		var errorMessage = ErrorLogger.ErrorMessage{ Error: geoErr, Message:"Geo database is not set up", Category: "setup", Priority: ErrorLogger.HIGH_PRIORITY}
		ErrorLogger.ReportError(errorMessage)
	}


	//Checking if all the environment variables exist
	if (dbName == "")|| (dbUser == "") || (awsRegion == "") ||
		(dbEndpoint == "")|| (dbPort == "") || (dbKey == ""){


		var errorMessage = ErrorLogger.ErrorMessage{ Error: errors.New("Environment keys missing"), Message:"Please insert all Environment keys.", Category: "setup" ,Priority: ErrorLogger.HIGH_PRIORITY}
		ErrorLogger.ReportErrorAndWait(errorMessage)
		log.Panic("Missing environment keys for Database Setup")
	}


	//// Setting up certificates
	rootCertPool := x509.NewCertPool() //NewCertPool returns a new, empty CertPool.
	pem, err := ioutil.ReadFile("rds-ca-bundle.pem") //reading the provided pem
	if err != nil {
		var errorMessage = ErrorLogger.ErrorMessage{ Error: err, Message:"Could not read AWS RDS Certificate. Please insert into root folder. " , Category: "setup" , Priority: ErrorLogger.HIGH_PRIORITY}
		ErrorLogger.ReportErrorAndWait(errorMessage)
		log.Panic("Could not read AWS RDS Certificate. Please insert into root folder")

	}

	//AppendCertsFromPEM attempts to parse a series of PEM encoded certificates.
	//pushing in the pem
	if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
		var errorMessage = ErrorLogger.ErrorMessage{ Error: errors.New("Failed to append PEM"), Message:"Please check the pem." , Category: "setup" , Priority: ErrorLogger.HIGH_PRIORITY}
		ErrorLogger.ReportErrorAndWait(errorMessage)
		log.Panic("Failed to append PEM.")
	}

	//setting up TLS Config
	mysql.RegisterTLSConfig("custom", &tls.Config{
		RootCAs: rootCertPool,
		InsecureSkipVerify: false,
	})





	//// Connecting
	log.Println("Connecting to the database..") //todo delete


	//setting up DNS String
	dnsStr = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=custom&parseTime=true&collation=utf8mb4_unicode_ci&charset=utf8mb4",
		dbUser, dbKey, dbEndpoint, dbPort, dbName,
	)

	//Opening connection
	database, err = sql.Open("mysql", dnsStr)
	if err != nil {

		var errorMessage = ErrorLogger.ErrorMessage{ Error: err, Message:"DATABASE CONNECTION FAILED." , Category: "setup" , Priority: ErrorLogger.HIGH_PRIORITY}
		ErrorLogger.ReportError(errorMessage)
		log.Print("DATABASE CONNECTION FAILED" + err.Error())
		failedSetup = true
	}

	//Testing connection
	err = database.Ping()
	if err != nil {

		isConnected = false
		var errorMessage = ErrorLogger.ErrorMessage{ Error: err, Message:"DATABASE SETUP FAILED.", Category: "setup" , Priority: ErrorLogger.HIGH_PRIORITY}
		ErrorLogger.ReportError(errorMessage)

		log.Print("DATABASE SETUP FAILED." + err.Error())
		failedSetup = true
	}else{
		isConnected = true
		log.Print("DATABASE IS READY.")
		failedSetup = false
	}


	go testConnection()
	defer waitGroup.Done()
}
