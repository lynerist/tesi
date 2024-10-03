const LAYOUT = {
    name: 'breadthfirst',
    directed: true,             // Makes the layout directed
    spacingFactor: 0.9,        // Reduce spacing to keep nodes closer
    nodeDistance: 50,          // Ideal distance between nodes at the same depth
    avoidOverlap: true,        // Attempt to prevent node overlap
    animate: true,             // Optional: enables animations
    animationDuration: 3000,    // Duration of the animation
    
}

var cy = cytoscape({
    container: document.getElementById('cy'),
    elements: [], // Inizialmente vuoto
    style: [ // Stili base per nodi e archi
        {
            selector: 'node',
            style: {
                'label': 'data(id)',  // Mostra l'id del nodo come etichetta
                'background-color': '#0074D9',
                'color': '#fff',
                'text-valign': 'center',
                'text-halign': 'center',
                'height': '50px',
                // Calcola la larghezza in base alla lunghezza della label
                'width': function(ele) {
                    const label = ele.data('id');
                    return Math.max(label.length * 10, 50); // 10px per carattere, larghezza minima di 40px
                },
            }
        },
        {
            selector: 'edge',
            style: {
                'line-color': '#FF4136',
                'width': 3,
                'target-arrow-shape': 'triangle',
                'target-arrow-color': '#FF4136',
                'curve-style': 'bezier'
            }
        }
    ],
    layout: LAYOUT
});

document.getElementById('fileInput').addEventListener('change', function(event) {
    const file = event.target.files[0]; // Ottieni il file selezionato
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
        reader.readAsText(file); // Leggi il file come testo
    }
});
