class GraphView {
   constructor() {
      this.svg = null;
      this.simulation = null;
      this.width = 0;
      this.height = 0;
      this.nodes = [];
      this.links = [];
      this.init();
   }

   init() {
      this.setupContainer();
      this.loadData();
   }

   setupContainer() {
      const container = document.getElementById("graphContainer");
      this.width = container.clientWidth;
      this.height = container.clientHeight;

      this.svg = d3
         .select("#graphContainer")
         .append("svg")
         .attr("width", this.width)
         .attr("height", this.height);
      const zoom = d3
         .zoom()
         .scaleExtent([0.1, 4])
         .on("zoom", (event) => {
            this.svg.select("g").attr("transform", event.transform);
         });

      this.svg.call(zoom);
      this.svg.append("g");
      this.svg
         .select("g")
         .append("rect")
         .attr("width", this.width * 2)
         .attr("height", this.height * 2)
         .attr("x", -this.width / 2)
         .attr("y", -this.height / 2)
         .attr("fill", "transparent");
   }

   async loadData() {
      try {
         const response = await fetch("/api/graph");
         if (!response.ok) {
            throw new Error("Failed to load graph data");
         }
         const data = await response.json();
         this.nodes = data.nodes || [];
         this.links = data.links || [];
         this.renderGraph();
      } catch (error) {
         console.error("Error loading graph data:", error);
         this.showError("Failed to load graph data");
      }
   }

   renderGraph() {
      document.getElementById("loading").style.display = "none";

      this.simulation = d3
         .forceSimulation(this.nodes)
         .force(
            "link",
            d3
               .forceLink(this.links)
               .id((d) => d.id)
               .distance(100)
         )
         .force("charge", d3.forceManyBody().strength(-50))
         .force("center", d3.forceCenter(this.width / 2, this.height / 2))
         .force(
            "collision",
            d3.forceCollide().radius((d) => this.getNodeRadius(d) + 3)
         );
      const link = this.svg
         .select("g")
         .selectAll(".link")
         .data(this.links)
         .enter()
         .append("line")
         .attr("class", "link")
         .attr("stroke", "#999")
         .attr("stroke-opacity", 0.6)
         .attr("stroke-width", 1);

      const node = this.svg
         .select("g")
         .selectAll(".node")
         .data(this.nodes)
         .enter()
         .append("g")
         .attr("class", "node")
         .call(
            d3
               .drag()
               .on("start", (event, d) => {
                  if (!event.active) this.simulation.alphaTarget(0.3).restart();
                  d.fx = d.x;
                  d.fy = d.y;
               })
               .on("drag", (event, d) => {
                  d.fx = event.x;
                  d.fy = event.y;
               })
               .on("end", (event, d) => {
                  if (!event.active) this.simulation.alphaTarget(0);
                  d.fx = null;
                  d.fy = null;
               })
         );
      node
         .append("circle")
         .attr("r", (d) => this.getNodeRadius(d))
         .attr("fill", (d) => this.getNodeColor(d))
         .attr("stroke", "#fff")
         .attr("stroke-width", 2);
      node
         .append("text")
         .attr("dx", (d) => this.getNodeRadius(d) + 5)
         .attr("dy", 4)
         .style("font-size", "12px")
         .style("font-family", "Arial, sans-serif")
         .style("font-weight", "bold")
         .text((d) => this.truncateText(d.name, 20));
      this.simulation.on("tick", () => {
         link
            .attr("x1", (d) => d.source.x)
            .attr("y1", (d) => d.source.y)
            .attr("x2", (d) => d.target.x)
            .attr("y2", (d) => d.target.y);

         node.attr("transform", (d) => `translate(${d.x},${d.y})`);
      });
   }

   getNodeRadius(d) {
      if (d.type === "root") {
         return 20;
      } else if (d.type === "category") {
         return 15;
      } else {
         return 8;
      }
   }

   getNodeColor(d) {
      if (d.type === "root") {
         return "#1f2937";
      } else if (d.type === "category") {
         return "#4f46e5";
      } else {
         return "#10b981";
      }
   }

   truncateText(text, maxLength) {
      if (text.length <= maxLength) return text;
      return text.substring(0, maxLength - 3) + "...";
   }

   showError(message) {
      const container = document.getElementById("graphContainer");
      container.innerHTML = `
            <div class="flex items-center justify-center h-full">
                <div class="text-center">
                    <i class="fas fa-exclamation-triangle text-4xl text-red-400 mb-4"></i>
                    <p class="text-red-600 dark:text-red-400">${message}</p>
                </div>
            </div>
        `;
   }
}
document.addEventListener("DOMContentLoaded", () => {
   new GraphView();
});
