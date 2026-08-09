import { ShoppingCart } from "lucide-react";

export default function Navbar({ cartCount = 0, onCartClick }: { cartCount?: number; onCartClick?: () => void }) {
    return (
        <nav className="glass sticky top-0 z-50 px-8 py-4 mb-8 flex justify-between items-center bg-black/40 backdrop-blur-md">
            <div className="flex items-center gap-2">
                <div className="w-8 h-8 rounded-full bg-gradient-to-tr from-blue-500 to-purple-500 flex items-center justify-center">
                    <span className="font-bold text-white text-xl leading-none">M</span>
                </div>
                <h1 className="text-xl font-bold tracking-tight">MicroStore</h1>
            </div>
            <div className="flex items-center gap-4">
                <button
                    onClick={onCartClick}
                    className="relative p-2 rounded-full hover:bg-white/10 transition-colors"
                >
                    <ShoppingCart className="w-6 h-6 text-gray-300" />
                    {cartCount > 0 && (
                        <span className="absolute top-0 right-0 w-5 h-5 bg-blue-500 text-white rounded-full text-xs font-bold flex items-center justify-center animate-pulse">
                            {cartCount}
                        </span>
                    )}
                </button>
            </div>
        </nav>
    );
}
