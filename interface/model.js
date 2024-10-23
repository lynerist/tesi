const PORT = "60124"
var MODEL 

var cy = cytoscape({
    container: document.getElementById('cy'),
    elements: [], 
    style: [ 
        {selector: 'node', style: NODE },
        {selector: 'edge', style: EDGE },
        {selector: ".tag", style:TAG },
        {selector: ".root", style:ROOT },
        {selector: ".dependencyAll", style:DEPENDENCYALL},
        {selector: ".dependencyNot", style:DEPENDENCYNOT},
        {selector: ".dependencyAny", style:DEPENDENCYANY},
        {selector: ".dependencyOne", style:DEPENDENCYONE}

    ],
    layout: LAYOUT
});

document.getElementById('fileInput').addEventListener('change', function(event) {
    const file = event.target.files[0]; 
    if (file) {
        loadJSON_GO(file)
    }
});

function resetCy(){
    cy.elements().remove(); // Rimuovi elementi esistenti (se necessario)
    cy.removeAllListeners()
}

function displayModel(model){
    resetCy()
    cy.add(model); // Aggiungi i nuovi elementi direttamente dal file

    let layout = LAYOUT
    layout.idealEdgeLength = 50 + 5*cy.edges().length
    console.log(layout.idealEdgeLength)
    cy.layout(layout).run();

    cy.elements().forEach(function(ele) {
        if (ele.group() == "edges"){
            popDeclaration(ele);
        }else{
            popVariables(ele)
        }
    });

    // EDGES HOVER LISTENERS
    cy.on('mouseover', 'edge', (event) => {
        if (event.target.tippy) {
            event.target.tippy.show();
        }
    });

    cy.on('mouseout', 'edge', (event) => {
        if (event.target.tippy) {
            event.target.tippy.hide();
        }
    });

    // NODES HOVER LISTENERS

    cy.on('mouseover', 'node', (event) => {
        if (event.target.tippy) {
            event.target.tippy.show();
        }
    });
    
    cy.on('mouseout', 'node', (event) => {
        if (event.target.tippy) {
            event.target.tippy.hide();
        }
    });

    // NODES CLICK LISTENERS

    cy.on("click", "node", (event) => {
        hundleActivation(event.target)
    })
    
    console.log("Grafo caricato con successo!");
}

function updateEdges(model){
    let edges = model.filter(element => element.group === 'edges');
    cy.edges().remove()
    cy.add(edges)
    cy.elements().forEach(function(ele) {
        if (ele.group() == "edges"){
            popDeclaration(ele);
        }
    });

    // EDGES HOVER LISTENERS
    cy.on('mouseover', 'edge', (event) => {
        if (event.target.tippy) {
            event.target.tippy.show();
        }
    });

    cy.on('mouseout', 'edge', (event) => {
        if (event.target.tippy) {
            event.target.tippy.hide();
        }
    });    
    console.log("Grafo aggiornato con successo!");
}

function popDeclaration(ele) {
    if(! ele.data("declaration")){
        ele.tippy = ""
        return
    }
    let ref = ele.popperRef(); 
    
    ele.tippy = tippy(ref, { 
        content: () => {
            let content = document.createElement('div');
            
            content.innerHTML = ele.data("declaration");
            content.innerHTML = content.innerHTML.replace(/\n/g, "<br />")
            return content;
        },
        trigger: 'manual' 
    });
}

function popVariables(ele) {
    if(Object.keys(ele.data("variables")).length == 0){
        ele.tippy = ""
        return
    }
    let ref = ele.popperRef(); 

    ele.tippy = tippy(ref, { 
        content: () => {
            let content = document.createElement('div');
            
            Object.entries(ele.data("variables")).forEach(([variableName, variableValue]) => {
                content.innerHTML = content.innerHTML + `${variableName}: <input type="text" id="${variableName}
                                                                            placeholder="${variableValue}" 
                                                                            onchange="updateVariable('${ele.data("id")}','${variableName}',value)"
                                                                            value="${variableValue}"/>\n`

            })
            content.innerHTML = content.innerHTML.replace(/\n/g, "<br />")
            return content;
        },
        interactive:true,
        arrow:true,
        allowHTML: true,
        trigger: 'manual' 
    });

    ele.tippy.popper.addEventListener('mouseenter', () => {
        ele.tippy.show();
    });
    ele.tippy.popper.addEventListener('mouseleave', () => {
        ele.tippy.hide();
    });
}

function activateUp(ele){
    parents = cy.nodes().edgesTo(ele).sources().filter((e)=> e.data('abstract'))
    parents.forEach((parent)=>parent.data("active", true) && activateUp(parent))
}

function unactivateDown(ele) {
    children = ele.data("abstract")?ele.edgesTo(cy.nodes()).targets():[]
    children.forEach((child)=>child.data("active", false) && unactivateDown(child))
}

function hundleActivation(ele){
    ele.data("active", !ele.data("active"))     //activate or unactivate node
    if (ele.data("active")){                    //activate parent node
        activateUp(ele)
    }else{
        unactivateDown(ele)
    }
    cy.style().update()
}

/* BACKEND */

function loadJSON_GO(file){
    let formData = new FormData();
    formData.delete("json")
    formData.append("json", file, file.name);

    fetch(`http://localhost:${PORT}/loadjson`, {  // Sostituisci con il tuo endpoint
        method: "POST",
        body: formData
    })
    .then(response => response.json())  // Converti la risposta in JSON (o altro formato, a seconda del server)
    .then(data => {
        displayModel(data)
        MODEL = data
    })
    .catch((error) => {
        console.error('Error:', error);
    });
}

function updateVariable(feature, name, value){
    let formData = new FormData();
    formData.delete("feature")
    formData.delete("name")
    formData.delete("value")
    formData.delete("model")

    formData.append("feature", feature)
    formData.append("name", name)
    formData.append("value", value)
    
    //formData.append("model", model)

    console.log(MODEL)

    fetch(`http://localhost:${PORT}/updateVariable`,{
        method: "POST",
        body: formData
    })
    .then(response => response.json())
    .then(data => {
        updateEdges(data)
    })
    .catch((error) => {
        console.error('Error:', error);
    });
}



// TODO MISSING REQUIREMENTS dead features
// TODO GLOBALS