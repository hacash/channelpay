/**
 * // 绑定的函数
 * ChangeAutoCollection(int)
 * PrequeryPayment(string,string)
 * 
 * // 调用的函数
 * Logout()
 * LogPrint(string, bool)
 * InitAccount(...)
 * UpdateBalance(...)
 * ShowPaymentError(string)
 * 
 */


/* 退出 */
function Logout(tip) {
    tip = tip || "You have logged out. Please log in again to collect money"
    alert("[Logout Attention] " + tip)
    window.close() // 关闭窗口
}

/* 日志输出 */
var logw = document.getElementById("logw");
function LogPrint(log, iserr) {
    var p = document.createElement("p");
    p.innerText = log;
    if(iserr){
        p.setAttribute("class", "e")
    }
    logw.appendChild(p);
    logw.scrollTop = logw.scrollHeight;
}


/* 初始化显示账户 */
var cid = document.getElementById("cid")
, addr = document.getElementById("addr")
;
function InitAccount(pcid, paddr) {
    cid.innerText = pcid;
    addr.innerText = paddr;
}

/* 更新余额显示 */
var blsamt = document.getElementById("blsamt")
, blscap = document.getElementById("blscap")
, blrun = document.getElementById("blrun")
, blanb = document.getElementById("blanb")
, nobill = document.getElementById("nobill")
, bdts = document.getElementById("bdts")
;
function UpdateBalance(bls, cap, reusenum, billno, billbodyhex) {
    // 数额
    blsamt.innerText = bls;
    blscap.innerText = cap;
    blrun.innerText = reusenum;
    blanb.innerText = billno;
    // 票据
    if(billbodyhex) {
        bdts.value = billbodyhex;
        bdts.style.display = "block";
        nobill.style.display = "none";
    }else{
        bdts.style.display = "none";
        nobill.style.display = "block";
    }
}

/* 显示支付错误 */
var payerr = document.getElementById("payerr")
;
function ShowPaymentError(errmsg) {
    payerr.innerText = errmsg;
}



/* 初始化运行 */
(function () {

    // 查看切确地址
    var ufadr = document.getElementById("ufadr")
    ;
    ufadr.onclick = function(){
        var ads = addr.innerText.split("_");
        addr.innerText = ads[0] + "_" + cid.innerText + "_" + ads[1];
        ufadr.style.display = "none"
    }

    // 自动全选复制票据
    var bdts = document.getElementById("bdts")
    ;
    bdts.onclick = function(){
        bdts.select();
        document.execCommand("Copy"); // 执行浏览器复制命令
    }

    /* 开关自动收款 */
    var clctt = document.getElementById("clctt")
    , clctt1 = document.getElementById("clctt1")
    , clctt2 = document.getElementById("clctt2")
    , clcttIsOpen = true
    ;
    clctt.onclick = async function() {
        if(clcttIsOpen){
            clcttIsOpen = false
            clctt.className = "tap"
            clctt1.style.display = "none"
            clctt2.style.display = "inline-block"
        }else{
            clcttIsOpen = true
            clctt.className = "tap open"
            clctt2.style.display = "none"
            clctt1.style.display = "inline-block"
        }
        // 回调绑定
        await ChangeAutoCollection(clcttIsOpen?1:0)
    }

    /* 点击开始支付 */
    var paybtn = document.getElementById("paybtn")
    , payaddr = document.getElementById("payaddr")
    , payamt = document.getElementById("payamt")
    , clearErr = function(){
        ShowPaymentError("") // 清除错误
    };
    payaddr.onchange = clearErr
    payamt.onchange = clearErr
    paybtn.onclick = async function() {
        var errmsg = await PrequeryPayment(payaddr.value, payamt.value)
        if(errmsg) {
            // 显示错误
            ShowPaymentError(errmsg)
            return
        }
        // 成功发起支付
        clearErr()
        
    }



})();



