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
        {selector: ".dependencyNot", style:DEPENDENCYNOT}

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
    cy.layout(LAYOUT).run(); // Applica un layout 
    console.log("Grafo caricato con successo!");
}

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
