const LAYOUT = {
    name: 'breadthfirst',
    name: "cose",
    directed: true,             // Makes the layout directed
    spacingFactor: 0.9,        // Reduce spacing to keep nodes closer
    nodeDistance: 50,          // Ideal distance between nodes at the same depth
    avoidOverlap: true,        // Attempt to prevent node overlap
    animate: true,             // Optional: enables animations
    animationDuration: 3000,    // Duration of the animation
}

const EDGE = {
    'line-color': '#FF4136',
    'width': 3,
    'target-arrow-shape': 'triangle',
    'target-arrow-color': '#FF4136',
    'opacity':0.7,
    'curve-style': 'bezier'
}

const NODE = {
    'label': function(ele) {
        let label = ele.data('id');
        return label.split("::")[0];
    },
    'background-color': '#1074D9',
    'color': '#fff',
    'text-valign': 'center',
    'text-halign': 'center',
    'height': '50px',
    'opacity':0.9,
    // Calcola la larghezza in base alla lunghezza della label
    'width': function(ele) {
        const label = ele.data('id');
        return Math.max(label.length * 10, 50); // 10px per carattere, larghezza minima di 40px
    }
}

const TAG = {
    'label': function(ele) {
        let label = ele.data('id');
        return label.split("::")[1];
    },
    'background-color': '#0000D9',
}

const ROOT = {
    label : 'Root',
    'background-color': '#002129',
}