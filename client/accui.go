package client

const AccUIhtmlContent = `
<html>
<head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1">
    <meta name="google" content="notranslate" />

    <title>Hacash channel pay client</title>
    
    <style>
    
*{
    border: none;
    padding: 0;
    margin: 0;
    font: 11px/1.5 tahoma,arial,'Hiragino Sans GB', '微软雅黑', 'sans-serif'; 
    list-style: none;
    text-decoration: none;
}

html, body, ol, ul{
    margin: 0; padding: 0;
}

.clear{
    clear: both;
}

.none{
    display: none;
}

button{
    display: inline-block;
    padding: 5px 12px;
    background: dodgerblue;
    color: #fff;
    font-size: 13px;
    cursor: pointer;
    border-radius: 4px;
}

/**********************************************/

.wbox {
    display: block;
    width: 960px;
    height: 640px;
    /* border: 1px #ddd dashed; */
}

.box {
    display: inline-block;
    vertical-align: top;
    width: 560px;
    height: 620px;
    padding: 10px;
}

.box.r {
    background-color: #f9f9f9;
    width: 360px;
}

h3.tt {
    font-size: 14px;
    padding: 20px 0 8px 0;
    color: slategrey;
}

.cid {
    padding: 10px 0;
    background-color: #efefef;
    border-radius: 4px;
}
.cid td {
    vertical-align: top;
}
#ufadr {
    display: inline-block;
    width: 20px;
    height: 20px;
    text-align: center;
    line-height: 20px;
    font-size: 14px;
    background-color: silver;
    color: white;
    cursor: pointer;
    border-radius: 100px;
    margin-left: 10px;
} 
.cid label {
    width: 130px;
    display: block;
    text-align: right;
    padding-right: 10px;
    color: #666;
}
.cid .addr {
    word-break: break-all;
}
.cid a {
    font-weight: bold;
    color: steelblue;
}
.cid a:hover {
    cursor: pointer;
    text-decoration: underline;
}

.blsw {
    padding: 10px 0;
    background-image: linear-gradient(to right, #005700, #7d7804);
    border-radius: 4px;
}
.blsw * {
    vertical-align: baseline;
}
.blsw label {
    width: 130px;
    padding-right: 10px;
    display: block;
    color: #ffffff99;
    text-align: right;
}
.blsw .amt {
    font-size: 20px;
    line-height: 20px;
    font-weight: bold;
    color: #fff;
}
.blsw .cap {
    color: goldenrod;
}


.logt {
    font-size: 15px;
    font-weight: bold;
    color: cadetblue;
    margin-bottom: 10px;
}
.logw {
    height: 590px;
    overflow-y: scroll;
}
.logw p{
    font-size: 12px;
    color: darkseagreen;
    line-height: 14px;
}

.logw p.e{
    color: indianred;
}

.bill {
    border: 1px #ddd solid;
    border-radius: 4px;
    padding: 10px;
}
.bill .meta {
    font-size: 13px;
    color: #666;
}
.bill .meta b {
    color: olivedrab;
    font-size: 13px;
    font-weight: bold;
}
.bill .bdts {
    width: 100%;
    height: 68px;
    line-height: 14px;
    font-size: 12px;
    border: 1px #ccc solid;
    margin-top: 10px;
    padding: 6px;
    cursor: text;
}
.bill #nobill {
    display: none;
    padding: 20px;
    color: gray;
    font-size: 13px;
}

/****************************************/
.clct span {
    height: 30px;
    line-height: 30px;
}
.clct span.open {
    color: mediumseagreen;
}
.clct span.close {
    color: indianred;
}

.clct .tap {
    transition: all 0.4s;
    vertical-align: middle;
    margin-right: 12px;
    display: inline-block;
    position: relative;
    height: 30px;
    border: 1px #bbb solid;
    border-radius: 100px;
    width: 62px;
    background-color: #e6e6e6;
    cursor: pointer;
}
.clct .tap b {
    transition: all 0.4s;
    display: block;
    width: 26px;
    height: 26px;
    position: absolute;
    top: 2px;
    left: 2px;
    background-color: white;
    border-radius: 100px;
    box-shadow: 1px 1px 2px #00000044;
}
.clct .tap:hover {
    border-color: #999;
}

.clct .tap.open {
    transition: all 0.5s;
    background-color: dodgerblue;
    box-shadow: 1px 2px 3px #00000055 inset;
    border-color: #fff;
}
.clct .tap.open b {
    left: 34px;
}

/****************************************/

.gopay {
    position: relative;
}
.gopay input {
    width: 100%;
    height: 36px;
    padding: 0 1%;
    line-height: 36px;
    border-radius: 4px;
    border: #bbb solid 1px;
}
.gopay input.amt {
    margin-top: 10px;
    width: 77%;
}
.gopay .trsbtn {
    position: absolute;
    right: 1px;
    top: 46px;
    height: 36px;
    font-weight: bold;
    border: 1px #1b76d0 solid;
}
.gopay .trsbtn:hover{
    cursor: pointer;
    background-color: #1b76d0;
}
.gopay .trsbtn.ban {
    background-color: #ddd !important;
    border-color: #aaa !important;
}
.gopay .err {

    padding: 10px;
    font-size: 12px;
    color: red;
}

/****************************************/

.dopay {    
    display: none;
    position: fixed;
    top: 0;
    left: 0;
    width: 960px;
    height: 640px;
    background-color: rgba(255,255,255,0.75);
}
.dopay .conb {
    background-color: #fff;
    box-shadow: 0 0 10px rgba(0,0,0,0.75);
    display: block;
    margin: 40px 100px;
    padding: 10px 20px;
    width: 700px;
    height: 540px;
    overflow-y: scroll;
}

.dopay .ttt {
    font-size: 16px;
    font-weight: bold;
    color: #999;
    border-bottom: 1px solid #bbb;
    padding-bottom: 6px;
}

.dopay p.note {
    padding: 10px 0;
    font-size: 12px;
    color: peru;
}
.dopay p.check {
    margin-top: 10px;
    font-size: 14px;
    color: seagreen;
    font-weight: bold;
    word-break: break-all;
}
.dopay .pil {
    display: block;
    padding: 10px;
    border: 2px dashed transparent;
    border-radius: 10px;
    font-size: 13px;
}
.dopay .pil input {
    margin-right: 10px;
}

.dopay .pil:hover {
    cursor: pointer;
    background-color: #dffcdf;
}

.dopay .pil.active {
    background-color: #87c586;
    border-color: darkgreen;
    font-weight: bold;
}

.dopay .btns {
    text-align: center;
}
.dopay .btns button {
    margin: 10px;
}
.dopay .cancel {
    color: #333;
    background-color: #bbb;
}
    </style>
    <link rel="stylesheet" href="pay.css" />

</head>
<body>

<div class="wbox"><div class="box l">

    <div class="blsw">
        <table>
            <tr><td><label>Channel Balance: </label></td><td><b class="amt" id="blsamt">ㄜ91,616,204:240</b></td></tr>
            <tr><td><label>Collection capacity: </label></td><td><b class="cap" id="blscap">ㄜ132:247</b></td></tr>
        </table>
    </div>

    <div class="cid">

        <table>
            <tr><td><label>Channel ID: </label></td><td><a href="https://explorer.hacash.org/channel/7ff377a442250bbd0de17ce8d2e6ba08" target="_blank" id="cid">7ff377a442250bbd0de17ce8d2e6ba08</a></td></tr>
            <tr><td><label>Collection address:</label></td><td><b class="addr" id=addr>12fEmV9HBRZfnfhypxmtW82TNHbCiHfkzU_HCPN1</b><b id=ufadr title="View target path address">↵</b></td></tr>
        </table>
    </div>

    <div class="bill">
        <h6 class="meta">[Reconciliation meta info] Reuse version: <b id="blrun">1</b>, Bill serial number: <b id="blanb">1234</b></h6>
        <p id="nobill">bill not exist yet.</p>
        <textarea readonly="true" class="bdts" id="bdts">b38aa1b37411567d4313de346864749d3dce9bf3b8ba157d272b41899312f6beb38aa1b37411567d4313de346864749d3dce9bf3b8ba157d272b41899312f6beb38aa1b37411567d4313de346864749d3dce9bf3b8ba157d272b41899312f6beb38aa1b37411567d4313de346864749d3dce9bf3b8ba157d272b41899312f6beb38aa1b37411567d4313de346864749d3dce9bf3b8ba157d272b41899312f6beb38aa1b37411567d4313de346864749d3dce9bf3b8ba157d272b41899312f6be</textarea>

    </div>


    <h3 class="tt">Collection:</h3>
    <div class="clct">
        <div class="tap open" id="clctt"><b></b></div>
        <span class="open" id="clctt1">✓ Enabled automatic collection</span>
        <span class="close" id="clctt2" style="display: none;">✗ Collection has been closed, you cannot receive funds</span>
    </div>

    <h3 class="tt">Payment:</h3>
    <div id="gopay" class="gopay">
        <input class="addr" id="payaddr" placeholder="Target channel collection address" />
        <input class="amt" id="payamt" placeholder="Amount: ㄜ125:246 or 1.25" />
        <button class="trsbtn" id="paybtn">Start transfer</button>
        <div class="err" id="payerr"></div>
    </div>
    


</div><div class="box r">

    <h5 class="logt">Log printing:</h5>
    <div class="logw" id="logw">
        <p>connect to server successfully!</p>
    </div>

</div></div>


<!-- 选择支付渠道 -->
<div id="dopay" class="dopay">
    <div class="conb">
        <h3 class="ttt">Select and confirm your payment</h3>
        <p class="check"></p>
        <p class="note">Note: Please select a payment path. If it fails, you can try to select another path to initiate payment again. Once the payment is successful, it is irrevocable.</p>
        <form id="slctps"> 
            <label class="pil"><input name="ptitem" type="radio" value="1" /></label> 
        </form>
        <div class="btns">
            <button class="cancel">Cancel</button>
            <button class="submit">Confirm payment</button>
        </div>
    </div>
</div>


<script>
/**
 * // 绑定的函数
 * ChangeAutoCollection(int)
 * PrequeryPayment(string,string) string
 * ConfirmPayment(pathselect) string
 * CancelPayment()
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




</script>
<script src="pay.js"></script>

</body>
</html>

`
