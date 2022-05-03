#! /bin/bash


function build(){
    echo "go build $1"
    cd $1
    go build $1
}


basedir=$(pwd)
echo "basedir=$basedir"

function read_dir(){
    for file in `ls $1` #注意此处这是两个反引号，表示运行系统命令
    do
        if [ -d $1"/"$file ] #注意此处之间一定要加上空格，否则会报错
        then
            if [ "$file" == "vendor" ]
            then 
                continue
            fi
            if [ "$file" == "node_modules" ]
            then 
                continue
            fi

            read_dir $1"/"$file
        else
            if [ "$file" == "main.go" ]
            then 
                #echo $1"/"$file #在此处处理文件即可
                build $1
            fi            
        fi
    done
}
#读取第一个参数
read_dir $basedir