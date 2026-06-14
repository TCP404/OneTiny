(function(){let e=document.createElement(`link`).relList;if(e&&e.supports&&e.supports(`modulepreload`))return;for(let e of document.querySelectorAll(`link[rel="modulepreload"]`))n(e);new MutationObserver(e=>{for(let t of e)if(t.type===`childList`)for(let e of t.addedNodes)e.tagName===`LINK`&&e.rel===`modulepreload`&&n(e)}).observe(document,{childList:!0,subtree:!0});function t(e){let t={};return e.integrity&&(t.integrity=e.integrity),e.referrerPolicy&&(t.referrerPolicy=e.referrerPolicy),e.crossOrigin===`use-credentials`?t.credentials=`include`:e.crossOrigin===`anonymous`?t.credentials=`omit`:t.credentials=`same-origin`,t}function n(e){if(e.ep)return;e.ep=!0;let n=t(e);fetch(e.href,n)}})();var e=`modulepreload`,t=function(e){return`/`+e},n={},r=function(r,i,a){let o=Promise.resolve();if(i&&i.length>0){let r=document.getElementsByTagName(`link`),s=document.querySelector(`meta[property=csp-nonce]`),c=s?.nonce||s?.getAttribute(`nonce`);function l(e){return Promise.all(e.map(e=>Promise.resolve(e).then(e=>({status:`fulfilled`,value:e}),e=>({status:`rejected`,reason:e}))))}o=l(i.map(i=>{if(i=t(i,a),i in n)return;n[i]=!0;let o=i.endsWith(`.css`),s=o?`[rel="stylesheet"]`:``;if(a)for(let e=r.length-1;e>=0;e--){let t=r[e];if(t.href===i&&(!o||t.rel===`stylesheet`))return}else if(document.querySelector(`link[href="${i}"]${s}`))return;let l=document.createElement(`link`);if(l.rel=o?`stylesheet`:e,o||(l.as=`script`),l.crossOrigin=``,l.href=i,c&&l.setAttribute(`nonce`,c),document.head.appendChild(l),o)return new Promise((e,t)=>{l.addEventListener(`load`,e),l.addEventListener(`error`,()=>t(Error(`Unable to preload CSS for ${i}`)))})}))}function s(e){let t=new Event(`vite:preloadError`,{cancelable:!0});if(t.payload=e,window.dispatchEvent(t),!t.defaultPrevented)throw e}return o.then(e=>{for(let t of e||[])t.status===`rejected`&&s(t.reason);return r().catch(s)})},i=`github.com/TCP404/OneTiny-cli/internal/gui.Service`,a=`/wails/runtime.js`,o=`0.1.0`,s=document.querySelector(`#app`);if(!s)throw Error(`missing #app`);var c=s,l={running:!1,stateLabel:`未运行`,address:``,config:{rootPath:`/Users/me/Downloads`,port:9090,maxLevel:0,isAllowUpload:!1,isSecure:!1},hasCredentials:!1,configPath:`~/Library/Application Support/tiny/config.yml`,accessLogPath:`~/Library/Application Support/tiny/access.log`,portRestartRequired:!1,lastError:``},u=[{value:``,label:`全部`},{value:`access`,label:`access`},{value:`download`,label:`download`},{value:`upload`,label:`upload`},{value:`login`,label:`login`},{value:`reject`,label:`reject`},{value:`error`,label:`error`}],d=l,f=[],p=ie(),m={},h=`panel`,g=``,_=!1,v=null,y=b();w();function b(){let e=async(e,...t)=>{let n=await S(()=>x(a));return n?(_=!1,await n(`${i}.${e}`,...t)):(_=!0,C(e,t))};return{GetStatus:()=>e(`GetStatus`),StartSharing:()=>e(`StartSharing`),StopSharing:()=>e(`StopSharing`),UpdateConfig:t=>e(`UpdateConfig`,t),SetCredentials:t=>e(`SetCredentials`,t),GetLogs:t=>e(`GetLogs`,t),ClearLogs:()=>e(`ClearLogs`),ChooseDirectory:t=>e(`ChooseDirectory`,t),ExportLogs:t=>e(`ExportLogs`,t),OpenConfigDir:()=>e(`OpenConfigDir`)}}async function x(e){return await r(()=>import(e),[])}async function S(e){try{return(await e()).Call?.ByName??null}catch{return null}}async function C(e,t){switch(await new Promise(e=>window.setTimeout(e,120)),e){case`GetStatus`:return d;case`StartSharing`:return d={...d,running:!0,stateLabel:`运行中`,address:q(d.config.port),lastError:``},d;case`StopSharing`:return d={...d,running:!1,stateLabel:`未运行`,address:``,lastError:``},d;case`UpdateConfig`:{let e=t[0];if(d.running&&e.port!==void 0&&e.port!==d.config.port&&!e.restartPort)throw d={...d,lastError:`修改端口需要确认并重启服务`},Error(d.lastError);let{restartPort:n,...r}=e,i={...d.config,...r};return d={...d,config:i,address:d.running?q(i.port):``,portRestartRequired:!1,lastError:``},d}case`ChooseDirectory`:return`/Users/me/Shared`;case`GetLogs`:return ae(p,t[0]??{});case`ClearLogs`:p=[];return;case`ExportLogs`:return`/Users/me/Downloads/onetiny-access.csv`;case`OpenConfigDir`:return;case`SetCredentials`:{let e=t[0];if(!e.username.trim())throw Error(`用户名不能为空`);if(!e.password.trim())throw Error(`密码不能为空`);if(e.password!==e.confirmPassword)throw Error(`两次输入的密码不一致`);return d={...d,config:{...d.config,isSecure:e.enableSecure?!0:d.config.isSecure},hasCredentials:!0,lastError:``},d}default:throw Error(`unknown mock method: ${e}`)}}async function w(){try{d=await y.GetStatus(),h===`logs`&&(f=await y.GetLogs(m)),g=W()}catch(e){g=$(e)}T()}function T(){let e=g||d.lastError;c.innerHTML=`
    <main class="shell">
      <section class="top-control" aria-label="共享状态">
        <div class="access-block">
          <div class="access-labels">
            <span class="label">访问地址</span>
            <span class="state ${d.running?`state-running`:``}">${Q(d.stateLabel)}</span>
          </div>
          <code>${Q(d.address||`服务未启动`)}</code>
        </div>
        <div class="top-actions">
          <button data-action="copy" ${d.address?``:`disabled`}>复制地址</button>
          <button class="primary" data-action="${d.running?`stop`:`start`}">
            ${d.running?`停止共享`:`启动共享`}
          </button>
        </div>
      </section>

      ${e?`<p class="notice">${Q(e)}</p>`:``}

      <header class="app-header">
        <div class="brand">
          <div class="mark">O</div>
          <div>
            <h1>OneTiny</h1>
            <p>局域网文件共享控制面板</p>
          </div>
        </div>
      </header>

      <nav class="tabs">
        ${E(`panel`,`控制面板`)}
        ${E(`security`,`安全设置`)}
        ${E(`logs`,`访问日志`)}
        ${E(`about`,`关于`)}
      </nav>

      <section class="content">
        ${D()}
      </section>
      ${ee()}
    </main>
  `,ne()}function E(e,t){return`<button class="tab ${h===e?`active`:``}" data-tab="${e}">${t}</button>`}function D(){switch(h){case`panel`:return O();case`security`:return k();case`logs`:return j();case`about`:return M()}}function O(){return`
    <div class="control-list">
      <label class="control-row directory-row">
        <span>共享目录</span>
        <input class="readonly-input" type="text" value="${Q(d.config.rootPath)}" readonly>
        <button type="button" data-action="choose-dir">选择</button>
      </label>

      <div class="control-row">
        <span>允许上传</span>
        <label class="switch">
          <input type="checkbox" data-toggle="upload" ${d.config.isAllowUpload?`checked`:``}>
          <span></span>
        </label>
      </div>

      ${A()}

      <label class="control-row">
        <span>端口</span>
        <input class="number-input" type="number" min="1" max="65535" step="1" value="${d.config.port}" data-number="port">
      </label>

      <label class="control-row">
        <span>最大访问层级</span>
        <input class="number-input" type="number" min="0" max="255" step="1" value="${d.config.maxLevel}" data-number="maxLevel">
      </label>
    </div>
  `}function k(){return`
    <div class="control-list">
      ${A()}
      <div class="control-row">
        <span>账号状态</span>
        <strong class="value-pill ${d.hasCredentials?`ok`:``}">
          ${d.hasCredentials?`已配置`:`未配置`}
        </strong>
      </div>
      <div class="control-row">
        <span>登录保护</span>
        <strong class="value-pill ${d.config.isSecure?`ok`:``}">
          ${d.config.isSecure?`已开启`:`已关闭`}
        </strong>
      </div>
    </div>
  `}function A(){return`
    <div class="control-row">
      <span>登录保护</span>
      <div class="inline-actions">
        <label class="switch">
          <input type="checkbox" data-toggle="secure" ${d.config.isSecure?`checked`:``}>
          <span></span>
        </label>
        <button type="button" data-action="credentials">账号设置</button>
      </div>
    </div>
  `}function j(){return`
    <form class="log-filters" aria-label="访问日志筛选">
      <label>
        <span>事件</span>
        <select name="event">
          ${u.map(e=>`
                <option value="${Q(e.value)}" ${e.value===(m.event??``)?`selected`:``}>
                  ${Q(e.label)}
                </option>
              `).join(``)}
        </select>
      </label>
      <label>
        <span>开始时间</span>
        <input name="since" type="datetime-local" value="${Q(X(m.since))}">
      </label>
      <label>
        <span>结束时间</span>
        <input name="until" type="datetime-local" value="${Q(X(m.until))}">
      </label>
      <div class="toolbar">
        <button type="button" data-action="refresh-logs">刷新</button>
        <button type="button" data-action="export-logs">导出 CSV</button>
        <button type="button" class="danger" data-action="clear-logs">清空</button>
      </div>
    </form>
    <div class="log-table">
      ${te()}
    </div>
  `}function M(){return`
    <div class="about-panel">
      <dl class="about">
        <dt>版本</dt>
        <dd>OneTiny GUI ${Q(o)}</dd>
        <dt>模式</dt>
        <dd>${Q(K(G()))}</dd>
        <dt>配置文件</dt>
        <dd>${Q(d.configPath||`-`)}</dd>
        <dt>访问日志</dt>
        <dd>${Q(d.accessLogPath||`-`)}</dd>
      </dl>
      <button data-action="open-config">打开配置目录</button>
    </div>
  `}function ee(){return v?`
    <dialog class="credential-dialog" aria-labelledby="credential-title">
      <form class="credential-form" method="dialog">
        <div class="dialog-header">
          <h2 id="credential-title">账号设置</h2>
          <button class="icon-button" type="button" data-action="close-credentials" aria-label="关闭">×</button>
        </div>
        ${v.error?`<p class="dialog-error">${Q(v.error)}</p>`:``}
        <label>
          <span>用户名</span>
          <input name="username" autocomplete="username" value="${Q(v.username)}">
        </label>
        <label>
          <span>密码</span>
          <input name="password" type="password" autocomplete="new-password" value="${Q(v.password)}">
        </label>
        <label>
          <span>确认密码</span>
          <input name="confirmPassword" type="password" autocomplete="new-password" value="${Q(v.confirmPassword)}">
        </label>
        <div class="dialog-actions">
          <button type="button" data-action="close-credentials">取消</button>
          <button class="primary" type="submit">保存</button>
        </div>
      </form>
    </dialog>
  `:``}function te(){return f.length===0?`<p class="empty">暂无访问日志</p>`:`
    <table>
      <thead>
        <tr>
          <th class="log-time">时间</th>
          <th class="log-ip">客户端 IP</th>
          <th class="log-method">方法</th>
          <th class="log-event">事件</th>
          <th class="log-path">路径</th>
          <th class="log-status">状态</th>
          <th class="log-result">结果</th>
        </tr>
      </thead>
      <tbody>
        ${f.map(e=>`
              <tr>
                <td class="log-time">${Q(re(e.time))}</td>
                <td>${Q(e.clientIP)}</td>
                <td>${Q(e.method||`-`)}</td>
                <td>${Q(e.event)}</td>
                <td class="log-path">${Q(e.path||`-`)}</td>
                <td>${Q(e.status?String(e.status):`-`)}</td>
                <td>${Q(e.result||`-`)}</td>
              </tr>
            `).join(``)}
      </tbody>
    </table>
  `}function ne(){c.querySelectorAll(`[data-tab]`).forEach(e=>{e.addEventListener(`click`,()=>{h=e.dataset.tab,w()})}),c.querySelector(`[data-action="start"]`)?.addEventListener(`click`,()=>{N(async()=>{d=await y.StartSharing(),g=W(),T()})}),c.querySelector(`[data-action="stop"]`)?.addEventListener(`click`,()=>{N(async()=>{d=await y.StopSharing(),g=W(),T()})}),c.querySelector(`[data-action="copy"]`)?.addEventListener(`click`,()=>{N(async()=>{d.address&&(await navigator.clipboard.writeText(d.address),g=`访问地址已复制`,T())})}),c.querySelector(`[data-action="choose-dir"]`)?.addEventListener(`click`,()=>{N(async()=>{let e=await y.ChooseDirectory(d.config.rootPath);e&&(d=await y.UpdateConfig({rootPath:e}),g=W(),T())})}),c.querySelectorAll(`[data-toggle="upload"]`).forEach(e=>{e.addEventListener(`change`,()=>{N(async()=>{d=await y.UpdateConfig({isAllowUpload:e.checked}),g=W(),T()})})}),c.querySelectorAll(`[data-toggle="secure"]`).forEach(e=>{e.addEventListener(`change`,()=>{P(e.checked)})}),c.querySelectorAll(`[data-action="credentials"]`).forEach(e=>{e.addEventListener(`click`,()=>{R(d.config.isSecure)})}),c.querySelectorAll(`[data-number]`).forEach(e=>{e.addEventListener(`change`,()=>{e.dataset.number===`port`?F(e):e.dataset.number===`maxLevel`&&I(e)})}),c.querySelector(`[data-action="open-config"]`)?.addEventListener(`click`,()=>{N(async()=>{await y.OpenConfigDir(),g=W(),T()})}),c.querySelector(`.log-filters`)?.addEventListener(`submit`,e=>{e.preventDefault(),N(async()=>{m=J(),f=await y.GetLogs(m),g=W(),T()})}),c.querySelector(`[data-action="refresh-logs"]`)?.addEventListener(`click`,()=>{N(async()=>{m=J(),f=await y.GetLogs(m),g=W(),T()})}),c.querySelector(`[data-action="export-logs"]`)?.addEventListener(`click`,()=>{N(async()=>{m=J();let e=await y.ExportLogs(m);g=e?`已导出到 ${e}`:W(),T()})}),c.querySelector(`[data-action="clear-logs"]`)?.addEventListener(`click`,()=>{window.confirm(`确定清空访问日志？`)&&N(async()=>{await y.ClearLogs(),f=[],g=W(),T()})}),c.querySelectorAll(`[data-action="close-credentials"]`).forEach(e=>{e.addEventListener(`click`,()=>{z()})}),c.querySelector(`.credential-form`)?.addEventListener(`submit`,e=>{e.preventDefault(),L()}),B()}function N(e){e().catch(e=>{g=$(e),T()})}async function P(e){if(e&&!d.hasCredentials){R(!0);return}N(async()=>{d=await y.UpdateConfig({isSecure:e}),g=W(),T()})}async function F(e){let t=U(e.value,1,65535,`端口`);if(t===null){T();return}if(t!==d.config.port){if(d.running&&!window.confirm(`修改端口需要重启共享服务，是否继续？`)){T();return}N(async()=>{d=await y.UpdateConfig({port:t,restartPort:d.running}),g=W(),T()})}}async function I(e){let t=U(e.value,0,255,`最大访问层级`);if(t===null){T();return}t!==d.config.maxLevel&&N(async()=>{d=await y.UpdateConfig({maxLevel:t}),g=W(),T()})}async function L(){if(!v)return;let e=H(`username`).trim(),t=H(`password`),n=H(`confirmPassword`),r=v.targetSecure;v={...v,username:e,password:t,confirmPassword:n,error:``};let i=V(e,t,n);if(i){v.error=i,g=``,T();return}N(async()=>{d=await y.SetCredentials({username:e,password:t,confirmPassword:n,enableSecure:r}),v=null,g=W(),T()})}function R(e){v={targetSecure:e,username:``,password:``,confirmPassword:``,error:``},g=``,T()}function z(){v=null,g=W(),T()}function B(){let e=c.querySelector(`.credential-dialog`);if(!e)return;let t=()=>{v&&(v=null,g=W(),T())};e.addEventListener(`cancel`,e=>{e.preventDefault(),t()}),e.addEventListener(`close`,t),e.open||e.showModal(),e.querySelector(`input[name="username"]`)?.focus()}function V(e,t,n){return e?t.trim()?t===n?``:`两次输入的密码不一致`:`密码不能为空`:`用户名不能为空`}function H(e){return c.querySelector(`.credential-form [name="${e}"]`)?.value??``}function U(e,t,n,r){let i=Number(e);return!Number.isInteger(i)||i<t||i>n?(g=`${r}必须在 ${t}-${n} 之间`,null):i}function W(){return _?`浏览器预览模式`:``}function G(){return _?`browser-preview`:`wails-desktop`}function K(e){return e===`browser-preview`?`浏览器预览模式`:`Wails 桌面运行时`}function q(e){return`http://127.0.0.1:${e}`}function J(){let e=c.querySelector(`.log-filters`);if(!e)return m;let t=new FormData(e),n=String(t.get(`event`)??``).trim(),r=Y(String(t.get(`since`)??``)),i=Y(String(t.get(`until`)??``)),a={};return n&&(a.event=n),r&&(a.since=r),i&&(a.until=i),a}function Y(e){if(!e)return null;let t=new Date(e);return Number.isNaN(t.getTime())?null:t.toISOString()}function X(e){if(!e)return``;let t=new Date(e);return Number.isNaN(t.getTime())?``:new Date(t.getTime()-t.getTimezoneOffset()*6e4).toISOString().slice(0,16)}function re(e){let t=new Date(e);return Number.isNaN(t.getTime())?e||`-`:new Intl.DateTimeFormat(void 0,{year:`numeric`,month:`2-digit`,day:`2-digit`,hour:`2-digit`,minute:`2-digit`,second:`2-digit`}).format(t)}function ie(){let e=Date.now(),t=t=>new Date(e-t*6e4).toISOString();return[{time:t(4),clientIP:`192.168.31.18`,method:`GET`,event:`access`,path:`/`,status:200,result:`ok`},{time:t(16),clientIP:`192.168.31.42`,method:`GET`,event:`download`,path:`/photos/2026/spring-trip/very-long-file-name-that-should-wrap-in-the-log-table-without-breaking-layout.jpg`,status:200,result:`sent`},{time:t(28),clientIP:`192.168.31.42`,method:`POST`,event:`upload`,path:`/uploads/report-final.pdf`,status:201,result:`created`},{time:t(44),clientIP:`192.168.31.9`,method:`POST`,event:`login`,path:`/login`,status:200,result:`authenticated`},{time:t(63),clientIP:`192.168.31.77`,method:`GET`,event:`reject`,path:`/private/<script>alert(1)<\/script>.txt`,status:403,result:`blocked`},{time:t(87),clientIP:`192.168.31.51`,method:`GET`,event:`error`,path:`/archive.zip`,status:500,result:`read failed`}]}function ae(e,t){let n=t.event?.trim(),r=Z(t.since),i=Z(t.until);return e.filter(e=>{let t=Z(e.time);return!(n&&e.event!==n||r!==null&&t!==null&&t<r||i!==null&&t!==null&&t>i)})}function Z(e){if(!e)return null;let t=new Date(e).getTime();return Number.isNaN(t)?null:t}function Q(e){return e.replace(/[&<>"']/g,e=>({"&":`&amp;`,"<":`&lt;`,">":`&gt;`,'"':`&quot;`,"'":`&#39;`})[e])}function $(e){return e instanceof Error?e.message:String(e)}