# FILE Cloud

<!-- linux -->
mkdir ~/root/go
mkdir $HOME/go/src/github.com/@github_name/
cd $HOME/go/src/github.com/@github_name/



## Prerequisites (Local)
XAMPP - https://sourceforge.net/projects/xampp/ 

Go    - https://go.dev/dl/

1. XAMPP -> start all service
2.  http://localhost/phpmyadmin/ -> user accounts -> new -> web(ndJMv9zrJw)
    2.1 import -> DB_export\newcc.sql

## Prerequisites (Server-Ubuntu)

1. Remove go https://go.dev/doc/manage-install#uninstalling)
  1.1 which go
  1.2 sudo rm -rf /usr/local/go
  1.3 sudo rm /etc/paths.d/go

2. Update go latest version (https://nextgentips.com/2021/12/23/how-to-install-go-1-18-on-ubuntu-20-04/)
  2.1 sudo apt update && apt upgrade -y
  2.2 curl -LO https://go.dev/dl/go1.19.2.linux-amd64.tar.gz
  2.3 sudo tar -C /usr/local -xzf go1.19.2.linux-amd64.tar.gz
  2.4 export GOPATH=$HOME/go

3. MySQL - (https://losst.pro/ustanovka-mysql-ubuntu-16-04)
  3.1 sudo apt update
  3.2 sudo apt install mysql-server mysql-client
  3.2.1 cd ~
  3.3 curl -LO https://dev.mysql.com/get/mysql-apt-config_0.8.24-1_all.deb
  3.4 sudo dpkg -i mysql-apt-config_0.8.24-1_all.deb -> Ok -> Ok
  3.5 sudo apt update
  3.6 sudo apt install mysql-server mysql-client -> Y

4. install list-extensions for vscode



5. start Mysql
  5.1 mysql -u root
  CREATE USER 'web'@'localhost' IDENTIFIED BY 'Todor1990///';
  CREATE DATABASE file_cloud;
  GRANT SELECT, INSERT ON file_cloud.*  TO 'web'@'localhost';


6. SQLTools(vscode extessions) -> Add New Connection -> name(alex_s); username(web); pass(ndJMv9zrJw) -> connect now


## Golang

1. Web app

  1.1 go run ./cmd/web -> localhost(mozilla)
  1.2 go run ./cmd/web -> localhost(mozilla)

3. Desktop app
4. Android app
