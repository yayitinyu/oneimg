import{r as c,j as G,o as ie,w as de,b as ce,c as l,d as o,f as r,k as V,e as x,n as u,F as D,p as H,t as p,i as F,q as O,y as ue,T as ge,z as me,x as L,s as J}from"./index-DQrT0z4C.js";import{e as pe}from"./error-BfqPR-Rg.js";const ve={class:"text-gray-800 dark:text-gray-200"},xe={class:"gallery-content container mx-auto px-4 py-8"},ye={key:0,class:"filter-bar mb-6 flex flex-wrap items-center justify-between gap-4"},fe={class:"role-filter flex items-center gap-3"},he={class:"role-buttons flex gap-1 p-1 rounded-full bg-gray-100 dark:bg-gray-800"},we={class:"view-toggle flex items-center gap-2"},be={key:1,class:"ml-4 text-sm text-primary font-medium"},ke={key:1,class:"loading-container flex flex-col items-center justify-center py-20"},_e={key:2,class:"images-container"},Ie={key:0,class:"images-grid grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4"},Ce=["onClick"],Me=["onClick"],$e={key:0,class:"ri-check-line text-sm"},Pe={class:"image-wrapper relative aspect-video overflow-hidden bg-gray-100 dark:bg-gray-900"},Le={class:"loading absolute inset-0 flex items-center justify-center z-0 text-slate-300"},je={class:"w-8 h-8 animate-spin",xmlns:"http://www.w3.org/2000/svg",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",style:{transform:"scaleX(-1) scaleY(-1)"}},ze=["src","alt"],Se={class:"image-info p-3"},Be={class:"image-filename font-medium text-sm truncate whitespace-nowrap overflow-hidden"},Te={class:"image-meta text-xs text-gray-500 dark:text-gray-400 mt-1"},Ee={class:"image-date text-xs text-gray-500 dark:text-gray-400 mt-1 truncate"},Ae={key:1,class:"columns-2 sm:columns-2 md:columns-3 lg:columns-4 gap-4 space-y-4"},De=["onClick"],He=["onClick"],Ze={key:0,class:"ri-check-line text-sm"},We={class:"relative overflow-hidden bg-gray-100 dark:bg-gray-900 rounded-2xl"},Ne={class:"loading absolute inset-0 flex items-center justify-center z-0 text-slate-300"},Re={class:"w-8 h-8 animate-spin",xmlns:"http://www.w3.org/2000/svg",fill:"none",viewBox:"0 0 24 24",stroke:"currentColor",style:{transform:"scaleX(-1) scaleY(-1)"}},Ue=["src","alt"],Ge={key:0,class:"flex items-center gap-2 text-gray-500"},Ve={key:1,class:"text-gray-400 text-sm"},Fe={key:3,class:"w-full py-8 text-center text-gray-400 text-sm"},Oe={key:4,class:"pagination flex flex-wrap items-center justify-center gap-2 py-8"},Je=["disabled"],qe={class:"page-numbers flex gap-1"},Qe=["onClick"],Xe=["disabled"],Ke={key:3,class:"empty-state flex flex-col items-center justify-center py-20 text-center"},Ye={class:"text-xl font-bold mb-2"},et={class:"text-gray-600 dark:text-gray-400 mb-6"},tt={key:0,class:"fixed bottom-6 right-6 z-50 flex flex-col items-end gap-3"},at={class:"floating-menu-badge bg-white dark:bg-gray-800 px-4 py-2 rounded-full shadow-lg text-sm font-medium text-center"},st={class:"floating-menu-buttons flex flex-col items-end gap-2"},rt=["title"],lt=["disabled"],ot=["disabled"],ct={__name:"Gallery",setup(nt){const w=e=>e?e.startsWith("http://")||e.startsWith("https://")||e.startsWith("//")?e:typeof window<"u"?window.location.origin+e:e:"",q=e=>e?e==="default"?"本地":e.charAt(0).toUpperCase()+e.slice(1):"未知",m=c([]),I=c(!1),g=c("grid"),d=c(1),y=c(1),Z=c(20),f=c("admin"),j=c(!1),h=c(!0),C=c(!1),z=c(null);let b=null;const v=c(!1),i=c([]),S=G(()=>m.value.length>0&&i.value.length===m.value.length),M=c(null),Q=()=>{v.value=!0,i.value=[]},W=()=>{v.value=!1,i.value=[]},$=e=>{const t=i.value.indexOf(e);t===-1?i.value.push(e):i.value.splice(t,1)},k=e=>i.value.includes(e),X=()=>{S.value?i.value=[]:i.value=m.value.map(e=>e.id)},K=async()=>{if(i.value.length===0)return;new PopupModal({title:"确认批量删除",content:`
            <div class="flex gap-3">
                <i class="ri-error-warning-line text-red-500 text-xl mt-1"></i>
                <div>
                    <p>确定要删除选中的 <strong>${i.value.length}</strong> 张图片吗？</p>
                    <p class="mt-1 text-secondary text-sm">图片将从存储中永久删除，无法恢复</p>
                </div>
            </div>
        `,buttons:[{text:"取消",type:"default",callback:t=>t.close()},{text:"确认删除",type:"danger",callback:async t=>{t.close(),await Y()}}],maskClose:!0}).open()},Y=async()=>{const e=Loading.show({text:`正在删除 ${i.value.length} 张图片...`,color:"#ff4d4f",mask:!0});let t=0,s=0;for(const a of i.value)try{(await fetch(`/api/images/${a}`,{method:"DELETE",headers:{Authorization:`Bearer ${localStorage.getItem("authToken")}`}})).ok?t++:s++}catch(n){console.error("删除图片错误:",n),s++}await e.hide(),s===0?Message.success(`成功删除 ${t} 张图片`):Message.warning(`删除完成：成功 ${t} 张，失败 ${s} 张`),W(),_()},ee=async()=>{if(i.value.length===0)return;const e=m.value.filter(s=>i.value.includes(s.id)),t=e.map(s=>w(s.url)).join(`
`);try{await navigator.clipboard.writeText(t),Message.success(`已复制 ${e.length} 个链接到剪贴板`)}catch(s){console.error("复制失败:",s),Message.error("复制失败")}},te=G(()=>{const e=[],t=Math.max(1,d.value-2),s=Math.min(y.value,d.value+2);for(let a=t;a<=s;a++)e.push(a);return e}),ae=me(),N=e=>{f.value!==e&&(f.value=e,d.value=1,_())},_=async()=>{I.value=!0;try{const e=new URLSearchParams({page:d.value,limit:Z.value,sort_by:"created_at",sort_order:"desc",role:f.value}),t=await fetch(`/api/images?${e}`,{headers:{Authorization:`Bearer ${localStorage.getItem("authToken")}`}});if(t.ok){const s=await t.json();m.value=s.data.images||[],y.value=s.data.total_pages||1,h.value=d.value<y.value}else{if(t.status===401){localStorage.removeItem("authToken"),ae.push("/login"),Message.error("登录已过期，请重新登录");return}throw new Error("加载图片失败")}}catch(e){console.error("加载图片错误:",e),Message.error("加载图片失败: "+e.message)}finally{I.value=!1,g.value==="masonry"&&L(()=>{T()})}},B=e=>{e>=1&&e<=y.value&&(d.value=e,_(),window.scrollTo({top:0,behavior:"smooth"}))},se=async()=>{if(!(C.value||!h.value||g.value!=="masonry")){C.value=!0;try{const e=d.value+1,t=new URLSearchParams({page:e,limit:Z.value,sort_by:"created_at",sort_order:"desc",role:f.value}),s=await fetch(`/api/images?${t}`,{headers:{Authorization:`Bearer ${localStorage.getItem("authToken")}`}});if(s.ok){const a=await s.json(),n=a.data.images||[];n.length>0&&(m.value=[...m.value,...n],d.value=e,y.value=a.data.total_pages||1),h.value=e<y.value}}catch(e){console.error("加载更多图片错误:",e)}finally{C.value=!1}}},T=()=>{b&&b.disconnect(),b=new IntersectionObserver(e=>{e.forEach(t=>{t.isIntersecting&&!C.value&&h.value&&se()})},{rootMargin:"100px"}),L(()=>{z.value&&b.observe(z.value)})},R=e=>{M.value=e;const t=new PopupModal({title:"图片预览",content:`
            <div class="image-preview-popup w-full max-w-[96vw] sm:max-w-5xl max-h-[85vh] flex flex-col overflow-hidden bg-white/85 dark:bg-dark-200/85 glass-card rounded-2xl">
                <!-- 顶部操作栏 -->
                <div class="preview-header bg-light-50/70 dark:bg-dark-300/70 pb-2 flex flex-col gap-2 px-3 sm:flex-row sm:flex-wrap sm:items-center sm:justify-between">
                    <div class="flex flex-col min-w-0 gap-1">
                        <h3 class="text-xs sm:text-sm font-medium truncate">${e.filename}</h3>
                        <p class="text-[11px] text-secondary truncate">${A(e.created_at)}</p>
                    </div>
                    <div class="flex gap-2 flex-wrap justify-end sm:justify-end w-full sm:w-auto">
                        <!-- 复制按钮 -->
                        <div class="relative z-100">
                            <button
                                class="halo-button-copy h-9 px-3 text-xs whitespace-nowrap flex items-center gap-1"
                                onclick="event.stopPropagation(); window.togglePreviewCopyMenu()"
                                title="复制链接"
                            >
                                <i class="ri-code-s-slash-line text-xs"></i>
                                <span>复制</span>
                            </button>
                            <!-- 复制下拉框 -->
                            <div
                                class="absolute left-1/2 sm:left-auto sm:right-0 top-full mt-1 w-32 bg-white/90 dark:bg-dark-200/90 rounded-xl shadow-2xl border border-white/40 dark:border-dark-100/60 backdrop-blur-xl z-101 transition-all duration-200 hidden opacity-0 translate-y-[-5px] -translate-x-1/2 sm:translate-x-0 z-[999]"
                                id="previewCopyDropdown"
                            >
                                <div class="p-1.5 space-y-1">
                                    <button
                                        class="w-full px-2 py-2 text-sm sm:text-xs text-gray-800 dark:text-light-100 hover:bg-light-100 dark:hover:bg-dark-300 rounded transition-colors duration-200 flex items-center justify-start gap-2 text-left"
                                        onclick="event.stopPropagation(); window.copyPreviewImageLink('url')"
                                    >
                                        <i class="ri-link text-primary"></i>
                                        <span class="font-semibold">URL</span>
                                    </button>
                                    <button
                                        class="w-full px-2 py-2 text-sm sm:text-xs text-gray-800 dark:text-light-100 hover:bg-light-100 dark:hover:bg-dark-300 rounded transition-colors duration-200 flex items-center justify-start gap-2 text-left"
                                        onclick="event.stopPropagation(); window.copyPreviewImageLink('html')"
                                    >
                                        <i class="ri-html5-line text-orange-500"></i>
                                        <span class="font-semibold">HTML</span>
                                    </button>
                                    <button
                                        class="w-full px-2 py-2 text-sm sm:text-xs text-gray-800 dark:text-light-100 hover:bg-light-100 dark:hover:bg-dark-300 rounded transition-colors duration-200 flex items-center justify-start gap-2 text-left"
                                        onclick="event.stopPropagation(); window.copyPreviewImageLink('markdown')"
                                    >
                                        <i class="ri-markdown-line text-blue-500"></i>
                                        <span class="font-semibold">MD</span>
                                    </button>
                                    <button
                                        class="w-full px-2 py-2 text-sm sm:text-xs text-gray-800 dark:text-light-100 hover:bg-light-100 dark:hover:bg-dark-300 rounded transition-colors duration-200 flex items-center justify-start gap-2 text-left"
                                        onclick="event.stopPropagation(); window.copyPreviewImageLink('bbcode')"
                                    >
                                        <i class="ri-brackets-line text-purple-500"></i>
                                        <span class="font-semibold">BBCode</span>
                                    </button>
                                </div>
                            </div>
                        </div>
                        <!-- 下载按钮 -->
                        <button
                            class="halo-button halo-button-primary h-9 px-3 text-xs whitespace-nowrap flex items-center gap-1"
                            onclick="event.stopPropagation(); window.downloadPreviewImage()"
                        >
                            <i class="ri-download-fill text-xs"></i>
                            下载
                        </button>
                        <!-- 删除按钮 -->
                        ${j.value?`
                        <button
                            class="halo-button text-danger h-9 px-3 text-xs whitespace-nowrap flex items-center gap-1"
                            onclick="event.stopPropagation(); window.deletePreviewImage('${e.id}')"
                        >
                            <i class="ri-delete-bin-fill text-xs"></i>
                            删除
                        </button>
                        `:""}
                    </div>
                </div>
                
                <!-- 预览图片区域 -->
                <div class="max-h-[360px] flex-1 overflow-auto flex items-center justify-center">
                    <a 
                        class="spotlight min-w-full max-w-full min-h-[260px] block" 
                        href="${w(e.url)}" 
                        data-description="尺寸: ${e.width||"未知"}×${e.height||"未知"} | 大小: ${E(e.file_size||0)} | 上传日期：${A(e.created_at)}"
                    >
                        <div class="relative max-w-full w-fill max-h-[360px] min-h-[260px] rounded-lg overflow-hidden image-skeleton flex items-center justify-center">
                            <img 
                                src="${w(e.url)}"
                                alt="${e.filename}" 
                                class="max-w-full w-fill max-h-[360px] min-h-[260px] object-contain rounded-lg relative z-10 opacity-0"
                                onload="this.classList.add('image-fade-in'); this.classList.remove('opacity-0'); this.parentElement.classList.remove('image-skeleton')"
                                onerror="this.parentElement.classList.remove('image-skeleton'); this.classList.remove('opacity-0'); this.src='${pe}';"
                            />
                        </div>
                    </a>
                </div>
                
                <!-- 底部信息栏 -->
                <!-- 底部信息栏 -->
                <div class="pt-2 flex flex-wrap gap-2 text-xs text-secondary ml-1 px-1">
                    <div class="flex items-center gap-1.5">
                        <i class="ri-ruler-line w-3.5 text-center"></i>
                        尺寸: ${e.width||"未知"}×${e.height||"未知"}
                    </div>
                    <div class="flex items-center gap-1.5">
                        <i class="ri-image-line w-3.5 text-center"></i>
                        大小: ${E(e.file_size||0)}
                    </div>
                    <div class="flex items-center gap-1.5">
                        <i class="ri-hard-drive-3-line"></i>
                        存储: ${e.storage==="telegram"?"Telegram":q(e.storage)}
                    </div>
                </div>
            </div>
        `,type:"default",buttons:[{text:"确定",type:"default",callback:s=>{s.close(),delete window.togglePreviewCopyMenu,delete window.copyPreviewImageLink,delete window.downloadPreviewImage,delete window.deletePreviewImage}}],maskClose:!0,zIndex:1e4,maxHeight:"90vh"});window.togglePreviewCopyMenu=()=>{const s=document.getElementById("previewCopyDropdown");s&&(s.classList.contains("hidden")?(s.classList.remove("hidden","opacity-0","translate-y-[-5px]"),s.classList.add("block","opacity-100","translate-y-0")):(s.classList.add("hidden","opacity-0","translate-y-[-5px]"),s.classList.remove("block","opacity-100","translate-y-0")))},window.copyPreviewImageLink=s=>re(s),window.downloadPreviewImage=()=>le(),window.deletePreviewImage=async s=>{t.close(),await oe(s)},t.open()},re=async e=>{if(!M.value)return;const t=M.value,s=w(t.url);let a="";switch(e){case"url":a=s;break;case"html":a=`<img src="${s}" alt="${t.filename}" width="${t.width||""}" height="${t.height||""}">`;break;case"markdown":a=`![img](${s})`;break;case"bbcode":a=`[img]${s}[/img]`;break;default:a=s}try{await navigator.clipboard.writeText(a),Message.success(`已复制${e.toUpperCase()}格式链接`,{position:"top-center",zIndex:2e4})}catch{const P=document.createElement("textarea");P.value=a,document.body.appendChild(P),P.select(),document.execCommand("copy"),document.body.removeChild(P),Message.success(`已复制${e.toUpperCase()}格式链接`,{position:"top-center",zIndex:2e4})}finally{L(()=>{const n=document.getElementById("previewCopyDropdown");n&&(n.classList.add("hidden","opacity-0","translate-y-[-5px]"),n.classList.remove("block","opacity-100","translate-y-0"))})}},le=()=>{if(!M.value)return;const e=M.value,t=document.createElement("a");t.href=e.url,t.download=e.filename,document.body.appendChild(t),t.click(),document.body.removeChild(t),Message.success("下载已开始")},oe=async e=>{new PopupModal({title:"确认删除",content:`
      <div class="flex gap-3">
        <i class="fa fa-exclamation-triangle text-warning text-xl mt-1"></i>
        <div>
          <p>确定要删除这张图片吗？</p>
          <p class="mt-1 text-secondary text-sm">删除后无法恢复，请谨慎操作</p>
        </div>
      </div>
    `,buttons:[{text:"取消",type:"default",callback:s=>s.close()},{text:"确认删除",type:"danger",callback:async s=>{s.close(),await ne(e)}}],maskClose:!0}).open()},ne=async e=>{const t=Loading.show({text:"删除中...",color:"#ff4d4f",mask:!0});try{const s=await fetch(`/api/images/${e}`,{method:"DELETE",headers:{Authorization:`Bearer ${localStorage.getItem("authToken")}`}});if(s.ok)return Message.success("图片删除成功"),_(),!0;{const a=await s.json();throw new Error(a.message||"删除失败")}}catch(s){return console.error("删除图片错误:",s),Message.error("删除图片失败: "+s.message),!1}finally{await t.hide()}},U=e=>{e.target.src="data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZGRkIi8+PHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtZmFtaWx5PSJBcmlhbCIgZm9udC1zaXplPSIxNCIgZmlsbD0iIzk5OSIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZHk9Ii4zZW0iPuWbvueJh+WKoOi9veWksei0pTwvdGV4dD48L3N2Zz4="},E=e=>{if(!e)return"0 B";const t=1024,s=["B","KB","MB","GB"],a=Math.floor(Math.log(e)/Math.log(t));return parseFloat((e/Math.pow(t,a)).toFixed(2))+" "+s[a]},A=e=>e?new Date(e).toLocaleString("zh-CN"):"";return ie(()=>{JSON.parse(localStorage.getItem("userInfo")||"{}")?.isTourist==!0?f.value="guest":j.value=!0,_(),T()}),de(g,e=>{d.value=1,h.value=!0,_(),e==="masonry"&&L(()=>{T()})}),ce(()=>{delete window.togglePreviewCopyMenu,delete window.copyPreviewImageLink,delete window.downloadPreviewImage,delete window.deletePreviewImage,b&&b.disconnect()}),(e,t)=>{const s=ue("router-link");return o(),l("div",ve,[r("div",xe,[!I.value&&j.value?(o(),l("div",ye,[r("div",fe,[t[8]||(t[8]=r("span",{class:"text-sm text-gray-600 dark:text-gray-400"},"查看角色：",-1)),r("div",he,[r("button",{onClick:t[0]||(t[0]=a=>N("admin")),class:u(["px-4 py-1.5 text-sm rounded-full transition-all duration-300",[f.value==="admin"?"bg-white dark:bg-gray-700 text-primary shadow-sm font-medium":"text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300"]])}," 管理员 ",2),r("button",{onClick:t[1]||(t[1]=a=>N("guest")),class:u(["px-4 py-1.5 text-sm rounded-full transition-all duration-300",[f.value==="guest"?"bg-white dark:bg-gray-700 text-primary shadow-sm font-medium":"text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300"]])}," 游客 ",2)])]),r("div",we,[t[12]||(t[12]=r("span",{class:"text-sm text-gray-600 dark:text-gray-400"},"视图：",-1)),r("button",{onClick:t[2]||(t[2]=a=>g.value="grid"),class:u(["p-2 rounded hover:bg-gray-100 dark:hover:bg-gray-700 transition-all",{"text-primary":g.value==="grid"}])},[...t[9]||(t[9]=[r("i",{class:"ri-grid-fill"},null,-1)])],2),r("button",{onClick:t[3]||(t[3]=a=>g.value="masonry"),class:u(["p-2 rounded hover:bg-gray-100 dark:hover:bg-gray-700 transition-all",{"text-primary":g.value==="masonry"}])},[...t[10]||(t[10]=[r("i",{class:"ri-layout-masonry-line"},null,-1)])],2),v.value?(o(),l("span",be,"批量模式已开启")):(o(),l("button",{key:0,onClick:Q,class:"ml-4 px-3 py-1.5 text-sm rounded-lg border border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700 transition-all flex items-center gap-1"},[...t[11]||(t[11]=[r("i",{class:"ri-checkbox-multiple-line"},null,-1),r("span",{class:"hidden sm:inline"},"批量管理",-1),r("span",{class:"sm:hidden"},"批量",-1)])]))])])):x("",!0),I.value?(o(),l("div",ke,[...t[13]||(t[13]=[r("div",{class:"spinner w-10 h-10 border-4 border-gray-200 dark:border-gray-700 border-t-primary dark:border-t-primary rounded-full animate-spin mb-4"},null,-1),r("p",{class:"text-gray-600 dark:text-gray-400"},"加载中...",-1)])])):m.value.length>0?(o(),l("div",_e,[g.value==="grid"?(o(),l("div",Ie,[(o(!0),l(D,null,H(m.value,a=>(o(),l("div",{key:a.id,class:u(["image-card bg-white/80 dark:bg-gray-800/80 glass-card rounded-2xl shadow-md overflow-hidden hover:shadow-xl transition-all duration-300 cursor-pointer border border-white/50 dark:border-gray-700/60 relative",{"ring-2 ring-primary":v.value&&k(a.id)}]),onClick:n=>v.value?$(a.id):R(a)},[v.value?(o(),l("div",{key:0,class:"absolute top-2 right-2 z-10",onClick:J(n=>$(a.id),["stop"])},[r("div",{class:u(["w-6 h-6 rounded-full border-2 flex items-center justify-center transition-all",k(a.id)?"bg-primary border-primary text-white":"bg-white/90 dark:bg-gray-800/90 border-gray-300 dark:border-gray-600"])},[k(a.id)?(o(),l("i",$e)):x("",!0)],2)],8,Me)):x("",!0),r("div",Pe,[r("p",{class:u(["image-role text-xs mt-1 px-2 py-0.5 rounded inline-block absolute left-[15px] top-[5px] z-[999]",[a.user_id=="1"?"bg-pink-100 text-pink-800 dark:bg-pink-900 dark:text-pink-200":"bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200"]])},p(a.user_id=="1"?"管理员":"游客"),3),r("div",Le,[(o(),l("svg",je,[...t[14]||(t[14]=[r("path",{"stroke-linecap":"round","stroke-linejoin":"round","stroke-width":"2",d:"M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"},null,-1)])]))]),r("img",{src:w(a.thumbnail||a.url),alt:a.filename,class:"image-thumbnail w-full h-full object-cover transition-transform duration-500 hover:scale-105 opacity-0",loading:"lazy",referrerpolicy:"no-referrer",onLoad:t[4]||(t[4]=n=>{n.target.classList.remove("opacity-0"),n.target.parentElement.querySelector(".loading").classList.add("hidden")}),onError:U},null,40,ze)]),r("div",Se,[r("p",Be,p(a.filename),1),r("p",Te,p(E(a.file_size))+" • "+p(a.width)+"×"+p(a.height),1),r("p",Ee,p(A(a.created_at))+" • "+p(a.user_id=="1"?"管理员":"游客"),1)])],10,Ce))),128))])):g.value==="masonry"?(o(),l("div",Ae,[(o(!0),l(D,null,H(m.value,a=>(o(),l("div",{key:a.id,class:u(["masonry-card break-inside-avoid overflow-hidden rounded-2xl shadow-md hover:shadow-xl transition-all duration-300 cursor-pointer relative",{"ring-2 ring-primary":v.value&&k(a.id)}]),onClick:n=>v.value?$(a.id):R(a)},[v.value?(o(),l("div",{key:0,class:"absolute top-2 right-2 z-10",onClick:J(n=>$(a.id),["stop"])},[r("div",{class:u(["w-6 h-6 rounded-full border-2 flex items-center justify-center transition-all",k(a.id)?"bg-primary border-primary text-white":"bg-white/90 dark:bg-gray-800/90 border-gray-300 dark:border-gray-600"])},[k(a.id)?(o(),l("i",Ze)):x("",!0)],2)],8,He)):x("",!0),r("div",We,[r("div",Ne,[(o(),l("svg",Re,[...t[15]||(t[15]=[r("path",{"stroke-linecap":"round","stroke-linejoin":"round","stroke-width":"2",d:"M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"},null,-1)])]))]),r("img",{src:w(a.thumbnail||a.url),alt:a.filename,class:"w-full h-auto object-cover opacity-0 transition-all duration-500",loading:"lazy",referrerpolicy:"no-referrer",onLoad:t[5]||(t[5]=n=>{n.target.classList.remove("opacity-0"),n.target.parentElement.querySelector(".loading").classList.add("hidden")}),onError:U},null,40,Ue)])],10,De))),128))])):x("",!0),g.value==="masonry"&&!I.value&&h.value?(o(),l("div",{key:2,ref_key:"loadMoreTrigger",ref:z,class:"w-full h-20 flex items-center justify-center"},[C.value?(o(),l("div",Ge,[...t[16]||(t[16]=[r("i",{class:"ri-loader-4-line animate-spin"},null,-1),r("span",null,"加载中...",-1)])])):(o(),l("div",Ve,"向下滚动加载更多"))],512)):x("",!0),g.value==="masonry"&&!h.value&&m.value.length>0?(o(),l("div",Fe," 已加载全部图片 ")):x("",!0),g.value==="grid"&&y.value>1?(o(),l("div",Oe,[r("button",{onClick:t[6]||(t[6]=a=>B(d.value-1)),disabled:d.value<=1,class:u(["page-btn px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 hover:bg-gray-100 dark:hover:bg-gray-700 transition-all text-sm",{"opacity-50 cursor-not-allowed":d.value<=1}])}," 上一页 ",10,Je),r("div",qe,[(o(!0),l(D,null,H(te.value,a=>(o(),l("button",{key:a,onClick:n=>B(a),class:u(["w-9 h-9 flex items-center justify-center rounded-lg border transition-all text-sm",[a===d.value?"bg-primary text-white border-primary":"border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 hover:bg-gray-100 dark:hover:bg-gray-700"]])},p(a),11,Qe))),128))]),r("button",{onClick:t[7]||(t[7]=a=>B(d.value+1)),disabled:d.value>=y.value,class:u(["page-btn px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 hover:bg-gray-100 dark:hover:bg-gray-700 transition-all text-sm",{"opacity-50 cursor-not-allowed":d.value>=y.value}])}," 下一页 ",10,Xe)])):x("",!0)])):(o(),l("div",Ke,[t[18]||(t[18]=r("div",{class:"empty-icon text-6xl mb-4 text-gray-400 dark:text-gray-600"},[r("i",{class:"ri-image-ai-line"})],-1)),r("h3",Ye,"暂无"+p(f.value==="admin"?"管理员":"游客")+"图片",1),r("p",et,[F(" 还没有上传任何"+p(f.value==="admin"?"管理员":"游客")+"图片， ",1),V(s,{to:"/",class:"text-primary hover:underline"},{default:O(()=>[...t[17]||(t[17]=[F("去上传一些吧",-1)])]),_:1})])]))]),V(ge,{name:"float-menu"},{default:O(()=>[v.value?(o(),l("div",tt,[r("div",at," 已选 "+p(i.value.length)+" 项 ",1),r("div",st,[r("button",{onClick:X,class:"floating-btn halo-button w-12 h-12 rounded-full flex items-center justify-center text-lg",title:S.value?"取消全选":"全选"},[r("i",{class:u(S.value?"ri-checkbox-indeterminate-line":"ri-checkbox-multiple-line")},null,2)],8,rt),r("button",{onClick:ee,disabled:i.value.length===0,class:"floating-btn halo-button halo-button-primary w-12 h-12 rounded-full flex items-center justify-center text-lg disabled:opacity-50",title:"复制链接"},[...t[19]||(t[19]=[r("i",{class:"ri-link"},null,-1)])],8,lt),r("button",{onClick:K,disabled:i.value.length===0,class:"floating-btn halo-button halo-button-danger w-12 h-12 rounded-full flex items-center justify-center text-lg disabled:opacity-50",title:"删除"},[...t[20]||(t[20]=[r("i",{class:"ri-delete-bin-line"},null,-1)])],8,ot),r("button",{onClick:W,class:"floating-btn halo-button w-12 h-12 rounded-full flex items-center justify-center text-lg",title:"取消"},[...t[21]||(t[21]=[r("i",{class:"ri-close-line"},null,-1)])])])])):x("",!0)]),_:1})])}}};export{ct as default};
