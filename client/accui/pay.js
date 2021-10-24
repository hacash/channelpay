/**
 * // 绑定的函数
 * ChangeAutoCollection(int)
 * PrequeryPayment(string,string) string
 * ConfirmPayment(pathselect) string
 * CancelPayment()
 * 
 * // 调用的函数
 * Logout()
 * ShowLogOnPrint(string, bool)
 * InitAccount(...)
 * UpdateBalance(...)
 * ShowPaymentError(string)
 * 
 */


/* 退出 */
function Logout(tip) {
    document.getElementById("lgagtip").style.display = "block";
    setTimeout(function (){
        tip = tip || "You have logged out. Please login again to collect money"
        alert("[Logout Attention] " + tip)
        window.close() // 关闭窗口
    }, 100)
}

/* 日志输出 */
var logw = document.getElementById("logw")
, logbg = document.getElementById("logbg")
;
function noticeLog() {
    // 吸引注意
    logbg.className = "flicker"
    setTimeout(function(){
        logbg.className = ""
    }, 1600)
}
function ShowLogOnPrint(log, isok, iserr) {
    var p = document.createElement("p");
    p.innerHTML = log;
    if(iserr){
        p.setAttribute("class", "e")
    }else if(isok) {
        p.setAttribute("class", "ok")
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
, payaddr = document.getElementById("payaddr")
, payamt = document.getElementById("payamt")
;
function ShowPaymentError(errmsg) {
    payerr.innerText = errmsg;
}

/* 选择支付渠道、选择支付路径 */
var dopay = document.getElementById("dopay")
, slctps = document.getElementById("slctps")
, slpcheck = dopay.getElementsByClassName("check")[0]
, slpcancel = dopay.getElementsByClassName("cancel")[0]
, slpsubmit = dopay.getElementsByClassName("submit")[0]
, slpvalue = 0
;
slpcancel.onclick = function(){
    dopay.style.display = "none" // 关闭窗口
    CancelPayment() // 取消支付
}
slpsubmit.onclick = async function(){
    // 确认支付
    if(slpvalue == 0){
        return alert("Please select payment path")
    }
    // alert("发起支付！" + slpvalue)
    var err = await ConfirmPayment(slpvalue)
    if(err) {
        return alert("Do payment error: " + err)
    }
    // 成功发起支付
    dopay.style.display = "none" // 关闭窗口
    noticeLog() // 吸引目光到日志
}
function SelectPaymentPaths(noteinfo, paths) {
    slpvalue = 0 // 重置
    dopay.style.display = "block"
    var itemshtml = ""
    for(var i in paths){
        var v = parseInt(i) + 1
        , one = paths[i];
        itemshtml += '<label class="pil"><input name="ptitem" type="radio" value="'+v+'" />'+one+'</label>';
    }
    slctps.innerHTML = itemshtml; // 填充
    slpcheck.innerText = "Check: " + noteinfo;
    var items = slctps.getElementsByClassName("pil")
    , clearActives = function(){
        for(var i in items){
            items[i].className = "pil"
        }
    }
    var vanum = 1
    for(var i in items){
        (function(vn){
            items[i].onclick = function(){
                clearActives()
                this.className = "pil active"
                slpvalue = vn
            }
        })(vanum)
        vanum++
    }
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
    , clearErr = function(){
        ShowPaymentError("") // 清除错误
    };
    payaddr.onchange = clearErr
    payamt.onchange = clearErr
    paybtn.onclick = async function() {
        if(paybtn.className.indexOf("ban") > 0){
            return
        }
        paybtn.className = "trsbtn ban"
        setTimeout(function(){
            paybtn.className = "trsbtn" // 按钮状态回退
        }, 2000)
        var errmsg = await PrequeryPayment(payaddr.value, payamt.value)
        if(errmsg) {
            // 显示错误
            ShowPaymentError(errmsg)
            return
        }
        // 成功发起支付
        clearErr()
        // 回退状态
        paybtn.className = "trsbtn"
    }


})();



