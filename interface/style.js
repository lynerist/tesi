const LAYOUT = {
    name: "cose-bilkent",
    directed: true,             // Makes the layout directed
    spacingFactor: 0.9,        // Reduce spacing to keep nodes closer
    nodeDistance: 20,          // Ideal distance between nodes at the same depth
    idealEdgeLength: 250,
    avoidOverlap: true,        // Attempt to prevent node overlap
    animate: true,             // Optional: enables animations
    animationDuration: 3000,    // Duration of the animation
}

const EDGE = {
    'line-color': '#000000',
    'width': 3,
    'target-arrow-shape': 'triangle-backcurve',
    'target-arrow-color': '#000000',
    'opacity': function(ele){
        return ele.source().data("active") ? 0.95 : 0.3
    },
    'curve-style': 'bezier'
}

const NODE = {
    'label': function(ele) {
        return ele.data("id").split("::")[0];          //Remove node incremental id
    },
    'background-color': function(ele){
        return ele.data("deadFeature")?'#F94449':'#1074D9'
    },
    'color': '#fff',
    'text-valign': 'center',
    'text-halign': 'center',
    'height': '50px',
    'opacity': function(ele){
        return ele.data("active") && !ele.data("deadFeature") ? 0.9 : 0.2          //Higher opacity for active nodes  
    },
    'width': function(ele) {
        return Math.max(ele.data("id").length * 10, 50); // 10px for each char, min length 50px
    }
}

const TAG = {
    'label': function(ele) {
        return ele.data("id").split("::")[1];                //Remove node incremental id
    },
    'background-color': '#0000D9',
}

const ROOT = {
    label : 'Root',
    'background-color': '#002129',
}

var COLORS = [];
for (let i = 0; i < 300; i++) {
    COLORS.push("#" + Math.floor(Math.random() * 16777215).toString(16)); //16777215 is ffffff
}
COLORS = COLORS.filter((color) => ! /^#[0-3][0-9A-F][0-3][0-9A-F][0-3][0-9A-F]$/.test(color) && color.length==7) //REMOVE dark colors

const DEPENDENCYALL = {
    'line-color': function(ele) {
        return COLORS[ele.data("dependencyID")]
    },
    'target-arrow-color': function(ele) {
        return COLORS[ele.data("dependencyID")]
    },
    'target-arrow-shape': 'triangle',
    'width': 2,
}

const DEPENDENCYNOT = {
    'line-color': '#FF4136',
    'target-arrow-shape': 'tee',
    'width': 2,
    'line-style': 'dashed',
    'target-arrow-color': '#FF4136',
}

const DEPENDENCYANY = {
    'line-color': function(ele) {
        return COLORS[ele.data("dependencyID")+50]
    },
    'target-arrow-color': function(ele) {
        return COLORS[ele.data("dependencyID")+50]
    },
    'target-arrow-shape': 'circle-triangle',
    'width': 2,
}

const DEPENDENCYONE = {
    'line-color': function(ele) {
        return COLORS[ele.data("dependencyID")+100]
    },
    'target-arrow-color': function(ele) {
        return COLORS[ele.data("dependencyID")+100]
    },
    'target-arrow-shape': 'triangle-tee',
    'width': 2,
}
