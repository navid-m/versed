declare var d3: any;

class GraphView {
   svg: any = null;
   simulation: any = null;
   width: number;
   height: number;
   nodes: any[];
   links: any[];
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
         .on("zoom", (event: { transform: any }) => {
            this.svg!.select("g").attr("transform", event.transform);
         });

      this.svg!.call(zoom);
      this.svg!.append("g");
      this.svg!.select("g")
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
               .id((d: { id: any }) => d.id)
               .distance(100)
         )
         .force("charge", d3.forceManyBody().strength(-50))
         .force("center", d3.forceCenter(this.width / 2, this.height / 2))
         .force(
            "collision",
            d3.forceCollide().radius((d: any) => this.getNodeRadius(d) + 3)
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
               .on(
                  "start",
                  (
                     event: { active: any },
                     d: { fx: any; x: any; fy: any; y: any }
                  ) => {
                     if (!event.active)
                        this.simulation.alphaTarget(0.3).restart();
                     d.fx = d.x;
                     d.fy = d.y;
                  }
               )
               .on(
                  "drag",
                  (event: { x: any; y: any }, d: { fx: any; fy: any }) => {
                     d.fx = event.x;
                     d.fy = event.y;
                  }
               )
               .on(
                  "end",
                  (event: { active: any }, d: { fx: null; fy: null }) => {
                     if (!event.active) this.simulation.alphaTarget(0);
                     d.fx = null;
                     d.fy = null;
                  }
               )
         );
      node
         .append("circle")
         .attr("r", (d: any) => this.getNodeRadius(d))
         .attr("fill", (d: any) => this.getNodeColor(d))
         .attr("stroke", "#fff")
         .attr("stroke-width", 2);
      node
         .append("text")
         .attr("dx", (d: any) => this.getNodeRadius(d) + 5)
         .attr("dy", 4)
         .style("font-size", "12px")
         .style("font-family", "Arial, sans-serif")
         .style("font-weight", "bold")
         .text((d: { name: any }) => this.truncateText(d.name, 20));
      this.simulation.on("tick", () => {
         link
            .attr("x1", (d: { source: { x: any } }) => d.source.x)
            .attr("y1", (d: { source: { y: any } }) => d.source.y)
            .attr("x2", (d: { target: { x: any } }) => d.target.x)
            .attr("y2", (d: { target: { y: any } }) => d.target.y);

         node.attr(
            "transform",
            (d: { x: any; y: any }) => `translate(${d.x},${d.y})`
         );
      });
   }

   getNodeRadius(d: { type: string }) {
      if (d.type === "root") {
         return 20;
      } else if (d.type === "category") {
         return 15;
      } else {
         return 8;
      }
   }

   getNodeColor(d: { type: string }) {
      if (d.type === "root") {
         return "#1f2937";
      } else if (d.type === "category") {
         return "#4f46e5";
      } else {
         return "#10b981";
      }
   }

   truncateText(text: string, maxLength: number) {
      if (text.length <= maxLength) return text;
      return text.substring(0, maxLength - 3) + "...";
   }

   showError(message: string) {
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
