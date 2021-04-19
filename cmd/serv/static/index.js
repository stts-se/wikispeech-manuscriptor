"use strict";

const baseURL = window.location.protocol + '//' + window.location.host;
let wsBase = baseURL.replace(/^http/, "ws");  

let clientID = makeid(7);

let ws;



onload = function() {
    console.log("HEJ!");
    let url = wsBase + "/ws/"+clientID;
    ws = new WebSocket(url);
    ws.onopen = function() {
        listFilters();
        listBatches();
        listScripts();
        //blockSents([10,11]);
        fetchBatch("test_batch_1", 1, 20);
        fetchScript("test_script_1", 1, 20);
    }
    ws.onmessage = function(evt) {
	let resp = JSON.parse(evt.data);
	//console.log(resp);
	if ( resp.message_type === "db_stats" ) 
	    updateStats(resp);
	else if ( resp.message_type === "filters" ) 
	    updateFilterList(resp);
	else if ( resp.message_type === "batches" ) 
	    updateBatchList(resp);
	else if ( resp.message_type === "scripts" ) 
            updateScriptList(resp);
	else if ( resp.message_type === "fetched_script" ) 
            console.log(resp.message_type, resp);
	else if ( resp.message_type === "fetched_batch" ) 
            console.log(resp.message_type, resp);
	else if ( resp.message_type === "blocked_sents" ) 
            console.log(resp.message_type, resp);
	else if ( resp.message_type === "server_error" ) 
	    console.log("SERVER ERROR", resp);
 	else if ( resp.message_type != "keep_alive" ) 
	     console.log("unknown message from server", resp);
	
    }
}
	
document.getElementById("reload_stats").addEventListener("click", function() {
    loadStats();
});

function fetchScript(scriptName, pageSize, pageNumber) {
    let request = {
	'client_id': clientID,
    'message_type': 'fetch_script',
    'payload': JSON.stringify({
        'name': scriptName,
        'type': 'script',
        'page_size': pageSize,
        'page_number': pageNumber,
    }),
    };
    ws.send(JSON.stringify(request));
};

function fetchBatch(batchName, pageSize, pageNumber) {
    let request = {
	'client_id': clientID,
    'message_type': 'fetch_batch',
    'payload': JSON.stringify({
        'name': batchName,
        'type': 'batch',
        'page_size': pageSize,
        'page_number': pageNumber,
    }),
    };
    ws.send(JSON.stringify(request));
};
function blockSents(ids) {
    let request = {
	'client_id': clientID,
    'message_type': 'block_sents',
    'payload': JSON.stringify(ids),
    };
    ws.send(JSON.stringify(request));
};


function loadStats() {
    let request = {
	'client_id': clientID,
	'message_type': 'get_stats',
    };
    ws.send(JSON.stringify(request));
};

function listBatches() {
    let request = {
	'client_id': clientID,
	'message_type': 'list_batches',
    };
    ws.send(JSON.stringify(request));
};

function listScripts() {
    let request = {
	'client_id': clientID,
	'message_type': 'list_scripts',
    };
    ws.send(JSON.stringify(request));
};

function listFilters() {
    let request = {
	'client_id': clientID,
	'message_type': 'list_filters',
    };
    ws.send(JSON.stringify(request));
};

function updateScriptList(scripts) {
    console.log("scripts", scripts);
    let div = document.getElementById("scripts");
    div.innerHTML = "";
    let json = JSON.parse(scripts.payload);
    json.forEach( script=> {
        let tr = document.createElement("tr");
        let details = document.createElement("tr");
        details.style["display"] = "none";

        let td = document.createElement("td");
        td.setAttribute("colspan", "5");
        td.innerHTML = JSON.stringify(script);
        details.appendChild(td);

        td = document.createElement("td");
        td.innerHTML = "+";
        td.style["cursor"]="pointer";
        td.title = "Click for details";
        td.style["font-family"] = "courier, fixed-width";
        td.addEventListener("click", function(evt) {
            let target = evt.target;
            if (target.innerHTML === "+") {
                target.innerHTML = "-";
                details.style.removeProperty("display");
            } else {
                target.innerHTML = "+";
                details.style["display"]="none";
            }
        });
        tr.appendChild(td);
        
        td = document.createElement("td");
        td.innerHTML = script.options.script_name;
        tr.appendChild(td);

        td = document.createElement("td");
        td.innerHTML = script.output_size;
        tr.appendChild(td);

        td = document.createElement("td");
        td.innerHTML = script.timestamp;
    	tr.appendChild(td);
        
        div.appendChild(tr);
        div.appendChild(details);
    });
}

function updateBatchList(batches) {
    console.log("batches", batches);
    let div = document.getElementById("batches");
    div.innerHTML = "";
    let json = JSON.parse(batches.payload);
    json.forEach( batch=> {
        let tr = document.createElement("tr");
        let details = document.createElement("tr");
        details.style["display"] = "none";

        let td = document.createElement("td");
        td.setAttribute("colspan", "5");
        td.innerHTML = JSON.stringify(batch);
        details.appendChild(td);

        td = document.createElement("td");
        td.innerHTML = "+";
        td.style["cursor"]="pointer";
        td.title = "Click for details";
        td.style["font-family"] = "courier, fixed-width";
        td.addEventListener("click", function(evt) {
            let target = evt.target;
            if (target.innerHTML === "+") {
                target.innerHTML = "-";
                details.style.removeProperty("display");
            } else {
                target.innerHTML = "+";
                details.style["display"]="none";
            }
        });
        tr.appendChild(td);
        
        td = document.createElement("td");
        td.innerHTML = batch.batch_name;
        tr.appendChild(td);

        td = document.createElement("td");
        td.innerHTML = batch.output_size;
        tr.appendChild(td);

        td = document.createElement("td");
        td.innerHTML = batch.timestamp;
        tr.appendChild(td);
        
        div.appendChild(tr);
        div.appendChild(details);
    });
}

function updateFilterList(filters) {
    let div = document.getElementById("filters");
    div.innerHTML = "";
    let json = JSON.parse(filters.payload);
    json.unshift("Output batch name");
    json.unshift("Target size");
    json.forEach( filter => {
	    let tr = document.createElement("tr");
	    let td1 = document.createElement("td1");
	    td1.innerHTML = filter;
	    let td2 = document.createElement("td");
	    td2.innerHTML = "<input/>";
	    tr.appendChild(td1);
	    tr.appendChild(td2);
	    div.appendChild(tr);
    });
}

function updateStats(stats) {
    let dbStats = document.getElementById("db_stats");
    dbStats.innerHTML = "";
    let json = JSON.parse(stats.payload);
    Object.keys(json).forEach((key, i) => {
	let tr = document.createElement("tr");
	let td1 = document.createElement("td1");
	td1.innerHTML = key;
	let td2 = document.createElement("td");
	td2.innerHTML = json[key];
	tr.appendChild(td1);
	tr.appendChild(td2);
	dbStats.appendChild(tr);
    });    
}

// https://stackoverflow.com/questions/1349404/generate-random-string-characters-in-javascript
function makeid(length) {
   var result           = '';
   var characters       = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
   var charactersLength = characters.length;
   for ( var i = 0; i < length; i++ ) {
      result += characters.charAt(Math.floor(Math.random() * charactersLength));
   }
   return result;
}
