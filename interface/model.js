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
        {selector: ".dependencyOne", style:DEPENDENCYONE},
    ],
    layout: LAYOUT,
    minZoom: 0.05
});

document.getElementById('fileInput').addEventListener('change', function(event) {
    const file = event.target.files[0]; 
    if (file) {
        handleJSONloading(file)
    }
});

function resetCy(){
    cy.elements().remove(); // Rimuovi elementi esistenti (se necessario)
    cy.removeAllListeners()
}

function displayModel(model){
    resetCy()
    cy.add(model); // Aggiungi i nuovi elementi direttamente dal file

    let dependencies = cy.edges().filter((e) => e.source().data("abstract")==undefined) //to apply the layout considering just the feature model tree
    cy.remove(dependencies)

    let layout = LAYOUT
    layout.idealEdgeLength = 50 + 3*cy.edges().length //for each edge 3px more in distance
    console.log(layout.idealEdgeLength)
    cy.layout(layout).run();
    cy.add(dependencies)
    

    cy.elements().forEach(function(ele) {
        if (ele.group() == "edges"){
            popDeclaration(ele);
        }else{
            popAttributes(ele)
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
        handleActivation(event.target, event.target.id())
    })
    
    console.log("Grafo caricato con successo!");
}

function updateNodesData(model){
    model.filter(element => element.group === 'nodes').forEach((node)=>{
        cy.getElementById(node.data.id).data("deadFeature", false)
        cy.getElementById(node.data.id).data(node.data)
    })
    cy.style().update()
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

function popAttributes(ele) {
    if(Object.keys(ele.data("variables")).length==0 && Object.keys(ele.data("globals")).length==0){
        ele.tippy = ""
        return
    }
    let ref = ele.popperRef(); 

    ele.tippy = tippy(ref, { 
        content: () => {
            let content = document.createElement('div');
            
            Object.entries(ele.data("variables")).forEach(([variableName, variableValue]) => {
                content.innerHTML = content.innerHTML + `${variableName}: <input type="text" id="${variableName}"
                                                                            placeholder="${variableValue}" 
                                                                            onchange="handleAttributeUpdate('${ele.data("id")}','${variableName}',value)"
                                                                            value="${variableValue}"/>\n`

            })

            Object.entries(ele.data("globals")).forEach(([globalName, globalValue]) => {
                content.innerHTML = content.innerHTML + `${globalName}: <input type="text" id="${globalName}"
                                                                            placeholder="${globalValue}" 
                                                                            onchange="handleAttributeUpdate('${ele.data("id")}','${globalName}',value, true)"
                                                                            value="${globalValue}"/>\n`

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

function colorGivenNodes(nodes){
    cy.nodes().forEach((node)=> node.data("active",false))
    nodes.forEach((id) => cy.getElementById(id).data("active",true))
    cy.style().update()
}

function handleActivation(ele, feature){
    let formData = new FormData();
    formData.set("feature", feature)

    fetch(`http://localhost:${PORT}/activation`, {  
        method: "POST",
        body: formData
    })
    .then(response => response.json())  
    .then(activeNodes => {
        colorGivenNodes(activeNodes)
        handleValidation()
    })
    .catch((error) => {
        console.error('Error:', error);
    });
}

/* BACKEND */

function translateGoJSONtoCytoscapeJSON(file){
    return file 
}

function handleJSONloading(file){
    let formData = new FormData();
    formData.delete("json")
    formData.append("json", file, file.name);

    fetch(`http://localhost:${PORT}/loadjson`, {  // Sostituisci con il tuo endpoint
        method: "POST",
        body: formData
    })
    .then(response => response.json())  
    .then(data => {
        displayModel(translateGoJSONtoCytoscapeJSON(data))
        MODEL = data
    })
    .catch((error) => {
        console.error('Error:', error);
    });
}

function handleAttributeUpdate(feature, name, value, isGlobal=false){
    let formData = new FormData();
    formData.delete("feature")
    formData.delete("name")
    formData.delete("value")
    formData.delete("model")

    formData.append("feature", feature)
    formData.append("name", name)
    formData.append("value", value)

    fetch(`http://localhost:${PORT}/updateAttribute?isglobal=${isGlobal}`,{
        method: "POST",
        body: formData
    })
    .then(response => response.json())
    .then(data => {
        updateEdges(data)
        updateNodesData(data)      

        //Update all poppers with that global
        if (isGlobal){
            cy.nodes().filter((e) => e.data('globals')[name]!= undefined).forEach((node) => {
                node.tippy.popper.querySelector("#\\"+name).value = value
            })
        }
    })
    .catch((error) => {
        console.error('Error:', error);
    });
}

function handleValidation () {
    fetch(`http://localhost:${PORT}/validation`,{
        method: "POST",
    })
    .then(response => response.json())
    .then(data => {
        let messages = ""
        Object.entries(data.invalids).forEach(pair =>{
            const [invalidFeature, requirements] = pair
            //ALL
            Object.keys(requirements.ALL).forEach((declaration)=>{
                messages += `<span class="invalidFeature">${invalidFeature.split("::")[0]}</span> is missing <span class="declaration">${declaration}</span>. `
                if (Object.keys(data.providers[declaration]).length>0){
                    messages += `Should be solved activating: `
                    Object.keys(data.providers[declaration]).forEach((provider)=>{
                        messages += `<span class="providerFeature">${provider.split("::")[0]}</span>, `
                    })
                    messages = messages.substring(0, messages.length-2) + "."
                }
                messages += "<br>"
            })

            //NOT
            Object.keys(requirements.NOT).forEach((declaration)=>{
                messages += `<span class="invalidFeature">${invalidFeature.split("::")[0]}</span> can't live with <span class="declaration">${declaration}</span>. `
                if (Object.keys(data.providers[declaration]).length>0){
                    messages += `Should be solved unactivating: `
                    Object.keys(data.providers[declaration]).forEach((provider)=>{
                        messages += `<span class="providerFeature">${provider.split("::")[0]}</span>, `
                    })
                    messages = messages.substring(0, messages.length-2) + "."
                }
                messages += "<br>"
            })

            //ANY
            if (requirements.ANY != null){
                requirements.ANY.forEach((group)=>{
                    messages += `<span class="invalidFeature">${invalidFeature.split("::")[0]}</span> is missing `
                    Object.keys(group).forEach((declaration)=>{
                        messages += `<span class="declaration">${declaration}</span>, `
                    })
                    messages = messages.substring(0, messages.length-2) + ". "

                    messages += `Should be solved activating: ` //If can't be activated this message should not be displayed but for now it is.
                    Object.keys(group).forEach((declaration)=>{
                        Object.keys(data.providers[declaration]).forEach((provider)=>{
                            messages += `<span class="providerFeature">${provider.split("::")[0]}</span>, `
                        })
                    })
                    messages = messages.substring(0, messages.length-2) + "."

                    messages += "<br>"
                })
            }


            //ONE
            if (requirements.ONE != null){
                requirements.ONE.forEach((group)=>{
                    messages += `<span class="invalidFeature">${invalidFeature.split("::")[0]}</span> must have only one of `
                    Object.keys(group).forEach((declaration)=>{
                        messages += `<span class="declaration">${declaration}</span>, `
                    })
                    messages = messages.substring(0, messages.length-2) + ". "

                    messages += `Should be solved activating only one of: ` //If can't be activated this message should not be displayed but for now it is.
                    Object.keys(group).forEach((declaration)=>{
                        Object.keys(data.providers[declaration]).forEach((provider)=>{
                            messages += `<span class="providerFeature">${provider.split("::")[0]}</span>, `
                        })
                    })
                    messages = messages.substring(0, messages.length-2) + "."

                    messages += "<br>"
                })
            }

        })
        document.getElementById("cy").classList.remove((messages=="")?"invalid":"valid")
        document.getElementById("cy").classList.add((messages=="")?"valid":"invalid")            
        document.getElementById("outMessages").innerHTML = messages
    })
    .catch((error) => {
        console.error('Error:', error);
    });
}


function handleVerboseValidationSwitch(){
    fetch(`http://localhost:${PORT}/verboseValidationSwitch`, {})
    .then(response => response.json())
    .then(data => {
        if(document.getElementById("validate").classList.contains("verbose")){
            document.getElementById("validate").classList.remove("verbose")
        }else{
            document.getElementById("validate").classList.add("verbose")
        }
        handleValidation()
    })
    .catch((error) => {
        console.error('Error:', error);
    });
}

function handleExporting(){
    fetch(`http://localhost:${PORT}/exporting`, {})
    .then(response => response.json())
    .then(data => {
        
    })
    .catch((error) => {
        console.error('Error:', error);
    });
}