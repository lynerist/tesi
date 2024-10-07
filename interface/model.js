var cy = cytoscape({
    container: document.getElementById('cy'),
    elements: [], 
    style: [ 
        {selector: 'node', style: NODE },
        {selector: 'edge', style: EDGE },
        {selector: ".tag", style:TAG },
        {selector: ".root", style:ROOT }
    ],
    layout: LAYOUT
});

document.getElementById('fileInput').addEventListener('change', function(event) {
    const file = event.target.files[0]; 
    if (file) {
        const reader = new FileReader();
        reader.onload = function(e) {
            const contents = e.target.result; // Contenuto del file
            try {
                const json = JSON.parse(contents); // Parse JSON
                cy.elements().remove(); // Rimuovi elementi esistenti (se necessario)
                cy.add(json); // Aggiungi i nuovi elementi direttamente dal file
                cy.layout(LAYOUT).run(); // Applica un layout 
                console.log("Grafo caricato con successo!");
            } catch (error) {
                console.error("Errore durante il parsing del JSON:", error);
            }
        };
        reader.readAsText(file); 
    }
});
