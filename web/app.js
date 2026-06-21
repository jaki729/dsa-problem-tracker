let PROBLEMS = [];
let progress = {}; // problemId -> status, from the Go backend
let TOPICS = [];

let activeTopic = 'ALL';
let activeFilter = 'all';
let searchTerm = '';

async function api(path, opts){
  const res = await fetch(path, { credentials: 'same-origin', headers: {'Content-Type':'application/json'}, ...opts });
  if(!res.ok) throw new Error(await res.text());
  return res.json();
}

async function boot(){
  try{
    const [problems, prog] = await Promise.all([
      api('/api/problems'),
      api('/api/progress'),
    ]);
    PROBLEMS = problems;
    progress = prog || {};
    TOPICS = [...new Set(PROBLEMS.map(p=>p.topic))];
    document.getElementById('metaInfo').innerHTML = `<b>${PROBLEMS.length}</b> problems`;
    render();
  }catch(e){
    document.getElementById('metaInfo').textContent = 'backend unreachable — is the Go server running?';
    document.getElementById('content').innerHTML = `<div class="empty">Couldn't load data from /api/problems. Run the Go server (<code>go run .</code>) and reload this page.</div>`;
  }
}

function getStatus(id){ return progress[id] || 'todo'; }

async function setStatus(id, status){
  progress = await api('/api/progress', { method:'POST', body: JSON.stringify({ problemId: id, status }) });
  render();
}

function cycle(status){ return status==='todo' ? 'solved' : status==='solved' ? 'revisit' : 'todo'; }

function counts(list){
  let solved=0, revisit=0;
  list.forEach(p=>{ const s=getStatus(p.id); if(s==='solved') solved++; else if(s==='revisit') revisit++; });
  return {solved, revisit, total:list.length};
}

function renderSidebar(){
  const el = document.getElementById('topicList');
  el.innerHTML = '';
  el.appendChild(topicRow('ALL','All topics', PROBLEMS));
  TOPICS.forEach(t=>{
    const list = PROBLEMS.filter(p=>p.topic===t);
    el.appendChild(topicRow(t, t, list));
  });
}
function topicRow(key, label, list){
  const c = counts(list);
  const pct = c.total ? Math.round((c.solved/c.total)*100) : 0;
  const row = document.createElement('div');
  row.className = 'topic-row' + (activeTopic===key ? ' active' : '');
  row.innerHTML = `<span class="tname">${label}</span><span class="tcount">${c.solved}/${c.total}</span><span class="tbar"><span style="width:${pct}%"></span></span>`;
  row.onclick = ()=>{ activeTopic = key; render(); };
  return row;
}

function matches(p){
  if(activeTopic!=='ALL' && p.topic!==activeTopic) return false;
  const s = getStatus(p.id);
  if(activeFilter!=='all' && s!==activeFilter) return false;
  if(searchTerm && !p.name.toLowerCase().includes(searchTerm)) return false;
  return true;
}

function render(){
  renderSidebar();
  const content = document.getElementById('content');
  content.innerHTML = '';

  const groups = activeTopic==='ALL' ? TOPICS : [activeTopic];
  let anyShown = false;

  groups.forEach(topic=>{
    const list = PROBLEMS.filter(p=>p.topic===topic).filter(matches);
    if(list.length===0) return;
    anyShown = true;
    const block = document.createElement('div');
    block.className = 'topic-block';
    const allInTopic = PROBLEMS.filter(p=>p.topic===topic);
    block.innerHTML = `<div class="topic-title"><span class="hash">#</span> ${topic} <span class="count">${counts(allInTopic).solved}/${allInTopic.length} solved</span></div>`;
    const table = document.createElement('table');
    const tbody = document.createElement('tbody');
    list.forEach(p=>{
      const tr = document.createElement('tr');
      const status = getStatus(p.id);
      tr.innerHTML = `
        <td class="idx">${p.idx}</td>
        <td class="name${status==='solved'?' name-solved':''}"><a href="${p.url}" target="_blank" rel="noopener">${p.name}</a></td>
        <td class="status"><span class="pill" data-status="${status}"><span class="dotmark"></span>${status}</span></td>
      `;
      tr.querySelector('.pill').onclick = ()=>{
        setStatus(p.id, cycle(getStatus(p.id)));
      };
      tbody.appendChild(tr);
    });
    table.appendChild(tbody);
    block.appendChild(table);
    content.appendChild(block);
  });

  if(!anyShown){
    content.innerHTML = '<div class="empty">no problems match your search/filter.</div>';
  }

  renderOverall();
}

function renderOverall(){
  const c = counts(PROBLEMS);
  document.getElementById('bigCount').innerHTML = `${c.solved} <span class="of">/ ${c.total} solved</span>${c.revisit? ` · ${c.revisit} to revisit`:''}`;
  document.getElementById('segSolved').style.width = (c.total? c.solved/c.total*100:0)+'%';
  document.getElementById('segRevisit').style.width = (c.total? c.revisit/c.total*100:0)+'%';
}

document.getElementById('search').addEventListener('input', e=>{
  searchTerm = e.target.value.trim().toLowerCase();
  render();
});

document.querySelectorAll('.fbtn').forEach(btn=>{
  btn.onclick = ()=>{
    document.querySelectorAll('.fbtn').forEach(b=>b.classList.remove('active'));
    btn.classList.add('active');
    activeFilter = btn.dataset.f;
    render();
  };
});

document.getElementById('resetBtn').onclick = async ()=>{
  if(confirm('Clear all saved progress for this session on the server?')){
    await api('/api/reset', { method:'POST' });
    progress = {};
    render();
  }
};

document.getElementById('menubtn').onclick = ()=>{
  document.getElementById('sidebar').classList.toggle('open');
};

boot();
