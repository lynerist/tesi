const PORT = "60124"

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

function displayModel(model){
    cy.elements().remove(); // Rimuovi elementi esistenti (se necessario)
    cy.add(model); // Aggiungi i nuovi elementi direttamente dal file

    let layout = LAYOUT
    layout.idealEdgeLength = 50 + 5*cy.edges().length
    console.log(layout.idealEdgeLength)
    cy.layout(layout).run();

    cy.elements().forEach(function(ele) {
        if (ele.group() == "edges"){
            makePopper(ele);
        }
    });

    cy.on('mouseover', '.dependency', (event) => {
        if (event.target.tippy) {
            event.target.tippy.show();
        }
    });

    cy.on('mouseout', '.dependency', (event) => {
        if (event.target.tippy) {
            event.target.tippy.hide();
        }
    });

    cy.on("click", "node", (event) => {
        activation(event.target)
    })
    
    console.log("Grafo caricato con successo!");
}

function makePopper(ele) {
    let ref = ele.popperRef(); // used only for positioning
    
    ele.tippy = tippy(ref, { // tippy options:
        content: () => {
            let content = document.createElement('div');
            content.innerHTML = ele.data("declaration");
            content.innerHTML = content.innerHTML.replace(/\n/g, "<br />")
            return content;
        },
        trigger: 'manual' 
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

function activation(ele){
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
    formData.append('json', file, file.name);

    fetch(`http://localhost:${PORT}/loadjson`, {  // Sostituisci con il tuo endpoint
        method: 'POST',
        body: formData,  // Qui passa il FormData con il file
    })
    .then(response => response.json())  // Converti la risposta in JSON (o altro formato, a seconda del server)
    .then(data => {
        displayModel(data)
    })
    .catch((error) => {
        console.error('Error:', error);
    });
}
