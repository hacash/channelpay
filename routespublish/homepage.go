package routespublish

func GetHomePageHtml() string {
	return `
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1, user-scalable=no">

    <title>Hacash Channel Payment Union / Hacash通道支付联盟</title>

    <style>
        * {
            font-family: Courier, "Courier New", monospace;
            margin: 0 auto; padding: 0;
            color: #fff;
            font-size: 16px;
        }
        .wpage {
            margin: auto;
            padding: 40px 0;
            max-width: 800px;
            background-color: #345;
            text-align: center;
            /*border: #ccc 1px solid;*/
            border-top: none;
        }
        .ptt {
            background: rgba(0,0,0,0.3);
            padding: 24px 0 20px 0;
            font-size: 34px;
            margin-bottom: 40px;
        }
        .wpage>p {
            margin: 10px 30px;
        }
        .members {
            text-align: center;
        }
        .members table {
            margin: auto;
        }
        .members td.r {
            text-align: right;
        }
        .members b {
            display: inline-block;
            font-size: 18px;
        }
        .members i {
            display: inline-block;
            width: 80px;
            text-align: center;
        }
        p.tt {
            margin: 20px 0;
            font-size: 16px;
            line-height: 32px;
            height: 32px;
            color: rgba(255,255,255,0.4);
        }
        p.line{
            display: block;
            margin: 30px 0; padding: 1px;
            border-bottom: rgba(255,255,255,0.4) solid 1px;
        }
    </style>

</head>
<body>

    <div class="wpage">
        <h1 class="ptt">Hacash Channel Payment Union<br>通道链网络支付联盟</h1>

        <p lang="en-US">Hacash Channel Payment Union (HCPU) is all types of service providers (not ordinary payment users) in the Hacash channel chain payment settlement network system, including payment service provider nodes, arbitration watchtowers, channel bill backup, channel signature machine developers, and channel wallets and so on with non-profit organizations jointly established by service providers voluntarily.</p>
        <p>The alliance has the following functions: 1. Design and determine the division of functional modules, interface definitions and communication protocols related to the payment network; 2. Abstract and develop open source code libraries such as core general functional modules and interface SDKs required by the payment network as a whole (such as sequence Sign communication machine, etc.); 3. Maintain and distribute the top-level node list and backbone routing table data of the payment network; 4. Arbitrate and collectively discuss other things between the cross-nodes of the channel chain payment network;</p>

        <p>Hacash Channel Payment Union (HCPU) 全称为“Hacash通道支付联盟”，是Hacash通道链支付结算网络体系内所有服务商（非普通支付用户），包括支付服务商节点、仲裁瞭望塔、通道票据备份、通道签名机研发商、通道钱包等等服务提供方所自愿联合成立的非营利性组织。</p>
        <p>联盟具备如下职能：一、设计和确定与支付网络相关的功能模块划分、接口定义和通信协议；二、抽象及开发支付网整体所需的例如序签通信等核心通用功能模块和接口SDK等开源代码库；三、维护及分发支付网络的顶级节点清单和骨干路由表数据；四、对通道链支付网路跨节点间的其它事务进行仲裁和集合商议；</p>

        <p class="tt">联盟成员 / Members</p>
        <div class="members">
            <table>
                <tr><td class="r"><b>HACorg</b></td><td><i>. . .</i></td><td><a href="https://hacash.org" target="_blank">hacash.org</a></td></tr>
                <tr><td class="r"><b>PayInst</b></td><td><i>. . .</i></td><td><a href="https://payinst.com" target="_blank">payinst.com</a></td></tr>
            </table>
        </div>

        <p class="line"></p>
        <p class="tt">联系方式或社交媒体 / Contact or social media</p>
        <p>Discord: <a href="https://discord.gg/nj5RD6z6JN" target="_blank">discord.gg/nj5RD6z6JN</a></p>
        <p>Email: <a href="mailto:hcpu@hacash.org">hcpu@hacash.org</a></p>
    </div>

</body>
</html>
`
}
