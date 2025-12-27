import matplotlib.pyplot as plt
import matplotlib.patches as patches
import networkx as nx

def save_plot(filename):
    plt.axis('off')
    plt.savefig(filename, bbox_inches='tight', dpi=150, pad_inches=0.1)
    plt.close()

def draw_handshake():
    fig, ax = plt.subplots(figsize=(12, 3))
    ax.set_xlim(0, 68)
    ax.set_ylim(0, 1)
    
    # Offsets and lengths
    # <len=19><"BitTorrent protocol"><8 reserved><20 info_hash><20 peer_id>
    # 0-1: len
    # 1-20: string
    # 20-28: reserved
    # 28-48: info_hash
    # 48-68: peer_id
    
    parts = [
        (0, 1, "Length\n(1 byte)", "#FF9999"),
        (1, 19, "String: 'BitTorrent protocol'\n(19 bytes)", "#99CCFF"),
        (20, 8, "Reserved\n(8 bytes)", "#FFFF99"),
        (28, 20, "Info Hash\n(20 bytes)", "#99FF99"),
        (48, 20, "Peer ID\n(20 bytes)", "#FFCC99")
    ]
    
    for start, width, label, color in parts:
        rect = patches.Rectangle((start, 0.2), width, 0.6, linewidth=1, edgecolor='black', facecolor=color)
        ax.add_patch(rect)
        ax.text(start + width/2, 0.5, label, ha='center', va='center', fontsize=10, fontweight='bold')
        
        # Add byte marker labels below
        ax.text(start, 0.1, str(start), ha='center', fontsize=8)
    
    ax.text(68, 0.1, "68", ha='center', fontsize=8)
    ax.set_title("Handshake Message Structure (68 Bytes)", fontsize=14)
    save_plot('docs/images/handshake_bytes.png')

def draw_pieces_blocks():
    fig, ax = plt.subplots(figsize=(10, 6))
    ax.set_xlim(0, 100)
    ax.set_ylim(0, 100)
    
    # File
    rect_file = patches.Rectangle((10, 60), 80, 20, linewidth=2, edgecolor='black', facecolor='#DDDDDD')
    ax.add_patch(rect_file)
    ax.text(50, 70, "Full File (e.g., 100 MB)", ha='center', va='center', fontsize=12, fontweight='bold')
    
    # Pieces
    for i in range(5):
        rect = patches.Rectangle((10 + i*16, 60), 16, 20, linewidth=1, edgecolor='black', facecolor='none')
        ax.add_patch(rect)
        ax.text(10 + i*16 + 8, 85, f"Piece {i}", ha='center', fontsize=10)

    # Zoom lines
    ax.plot([10 + 16*2, 10], [60, 40], color='black', linestyle='--')
    ax.plot([10 + 16*3, 90], [60, 40], color='black', linestyle='--')
    
    # One Piece zoomed in
    rect_piece = patches.Rectangle((10, 20), 80, 20, linewidth=2, edgecolor='black', facecolor='#99FF99')
    ax.add_patch(rect_piece)
    ax.text(50, 30, "Piece N (e.g., 256 KB)", ha='center', va='center', fontsize=12, fontweight='bold')
    
    # Blocks
    for i in range(4):
        rect = patches.Rectangle((10 + i*20, 20), 20, 20, linewidth=1, edgecolor='black', facecolor='none')
        ax.add_patch(rect)
        ax.text(10 + i*20 + 10, 15, f"Block {i}\n(16 KB)", ha='center', fontsize=9)

    ax.set_title("Files split into Pieces, Pieces split into Blocks", fontsize=14)
    save_plot('docs/images/pieces_blocks.png')

def draw_swarm_dynamics():
    G = nx.DiGraph()
    G.add_nodes_from(["Me", "Peer A", "Peer B", "Peer C"])
    
    pos = {
        "Me": (0.5, 0.5),
        "Peer A": (0.5, 0.8),
        "Peer B": (0.2, 0.2),
        "Peer C": (0.8, 0.2)
    }
    
    plt.figure(figsize=(6, 6))
    
    # Draw nodes
    nx.draw_networkx_nodes(G, pos, node_size=3000, node_color=['#FFFF99', '#99CCFF', '#99CCFF', '#99CCFF'], edgecolors='black')
    nx.draw_networkx_labels(G, pos, font_size=10, font_weight='bold')
    
    # Draw edges
    # Me <-> Peer A
    nx.draw_networkx_edges(G, pos, edgelist=[("Me", "Peer A")], connectionstyle="arc3,rad=0.1", arrows=True, edge_color='green', arrowstyle='-|>', width=2)
    nx.draw_networkx_edges(G, pos, edgelist=[("Peer A", "Me")], connectionstyle="arc3,rad=0.1", arrows=True, edge_color='green', arrowstyle='-|>', width=2)
    
    # Me -> Peer B (Optimistic Unchoke)
    nx.draw_networkx_edges(G, pos, edgelist=[("Me", "Peer B")], arrows=True, edge_color='blue', style='dashed', width=2)
    
    # Peer C -/-> Me (Choked)
    nx.draw_networkx_edges(G, pos, edgelist=[("Peer C", "Me")], arrows=True, edge_color='red', style='dotted', width=2, arrowstyle='-[')
    
    # Labels
    plt.text(0.35, 0.65, "Tit-for-Tat\n(Reciprocal)", color='green', fontsize=9, fontweight='bold', rotation=90)
    plt.text(0.35, 0.35, "Optimistic\nUnchoke", color='blue', fontsize=9, rotation=-45)
    plt.text(0.65, 0.35, "Choked", color='red', fontsize=9, rotation=45)
    
    plt.title("Swarm Dynamics & Choking", fontsize=14)
    plt.axis('off')
    plt.savefig('docs/images/swarm_dynamics.png', bbox_inches='tight', dpi=150)
    plt.close()

def draw_bitfield():
    fig, ax = plt.subplots(figsize=(10, 2))
    ax.set_xlim(0, 10)
    ax.set_ylim(0, 2)
    
    bits = [1, 0, 1, 1, 0, 0, 1, 0]
    
    for i, bit in enumerate(bits):
        color = "#99FF99" if bit == 1 else "#DDDDDD"
        rect = patches.Rectangle((i + 1, 0.8), 1, 1, linewidth=1, edgecolor='black', facecolor=color)
        ax.add_patch(rect)
        ax.text(i + 1.5, 1.3, str(bit), ha='center', va='center', fontsize=20, fontweight='bold')
        ax.text(i + 1.5, 0.5, f"Piece {i}", ha='center', fontsize=9)

    ax.text(0.5, 1.3, "Bits:", ha='center', fontsize=12, fontweight='bold')
    ax.text(5, 0.2, "1 = Have Piece, 0 = Don't Have", ha='center', fontsize=10, style='italic')
    
    ax.set_title("Bitfield Message (Bitmap of Available Pieces)", fontsize=14)
    save_plot('docs/images/bitfield.png')

if __name__ == "__main__":
    draw_handshake()
    draw_pieces_blocks()
    draw_swarm_dynamics()
    draw_bitfield()
    print("All diagrams generated successfully.")
