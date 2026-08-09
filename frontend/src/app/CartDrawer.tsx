'use client';

import { motion, AnimatePresence } from "framer-motion";
import { X, ShoppingCart, Trash2, CreditCard } from "lucide-react";
import { useState } from "react";
import { checkoutCart, removeFromCart } from "./actions";

interface CartItem {
    id: string;
    product_id: string;
    quantity: number;
    unit_price: number;
    product?: {
        name: string;
        category: string;
        image_url?: string;
    };
}

interface Cart {
    id: string;
    items: CartItem[];
    total_price: number;
    status: string;
}

interface CartDrawerProps {
    isOpen: boolean;
    onClose: () => void;
    cart: Cart | null;
    userId: string;
    onCartUpdate: (cart: Cart | null) => void;
}

export default function CartDrawer({ isOpen, onClose, cart, userId, onCartUpdate }: CartDrawerProps) {
    const [checkingOut, setCheckingOut] = useState(false);
    const [orderComplete, setOrderComplete] = useState<string | null>(null);
    const [removing, setRemoving] = useState<Record<string, boolean>>({});

    const items = cart?.items || [];
    const total = cart?.total_price || 0;

    const handleRemove = async (productId: string) => {
        setRemoving(prev => ({ ...prev, [productId]: true }));
        try {
            const updatedCart: any = await removeFromCart(userId, productId);
            onCartUpdate(updatedCart);
        } catch (error: any) {
            alert(error.message || 'Failed to remove item');
        } finally {
            setRemoving(prev => ({ ...prev, [productId]: false }));
        }
    };

    const handleCheckout = async () => {
        if (items.length === 0) return;
        setCheckingOut(true);
        try {
            const result: any = await checkoutCart(userId);
            setOrderComplete(result.order_id);
            onCartUpdate(null);
        } catch (error: any) {
            alert(error.message || "Checkout failed");
        } finally {
            setCheckingOut(false);
        }
    };

    return (
        <AnimatePresence>
            {isOpen && (
                <>
                    {/* Backdrop */}
                    <motion.div
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        exit={{ opacity: 0 }}
                        onClick={onClose}
                        className="fixed inset-0 bg-black/60 backdrop-blur-sm z-40"
                    />

                    {/* Drawer */}
                    <motion.div
                        initial={{ x: '100%' }}
                        animate={{ x: 0 }}
                        exit={{ x: '100%' }}
                        transition={{ type: 'spring', damping: 28, stiffness: 300 }}
                        className="fixed right-0 top-0 h-full w-full max-w-md bg-[#0f0f1a] border-l border-white/10 z-50 flex flex-col shadow-2xl"
                    >
                        {/* Header */}
                        <div className="flex items-center justify-between px-6 py-5 border-b border-white/10">
                            <div className="flex items-center gap-3">
                                <ShoppingCart className="w-5 h-5 text-blue-400" />
                                <h2 className="text-lg font-bold">Your Cart</h2>
                                {items.length > 0 && (
                                    <span className="bg-blue-500 text-white text-xs font-bold px-2 py-0.5 rounded-full">
                                        {items.reduce((s, i) => s + i.quantity, 0)}
                                    </span>
                                )}
                            </div>
                            <button onClick={onClose} className="p-2 rounded-full hover:bg-white/10 transition-colors">
                                <X className="w-5 h-5" />
                            </button>
                        </div>

                        {/* Body */}
                        <div className="flex-1 overflow-y-auto px-6 py-4 space-y-4">
                            {orderComplete ? (
                                <motion.div
                                    initial={{ opacity: 0, scale: 0.9 }}
                                    animate={{ opacity: 1, scale: 1 }}
                                    className="text-center py-16"
                                >
                                    <div className="text-6xl mb-4">🎉</div>
                                    <h3 className="text-2xl font-bold mb-2">Order Placed!</h3>
                                    <p className="text-gray-400 text-sm mb-1">Your order has been confirmed.</p>
                                    <p className="text-xs text-gray-500 font-mono">ID: {orderComplete}</p>
                                    <button
                                        onClick={() => { setOrderComplete(null); onClose(); }}
                                        className="mt-6 px-6 py-2 rounded-full bg-blue-500 hover:bg-blue-600 text-white font-semibold transition-colors"
                                    >
                                        Continue Shopping
                                    </button>
                                </motion.div>
                            ) : items.length === 0 ? (
                                <div className="text-center py-16 text-gray-500">
                                    <ShoppingCart className="w-12 h-12 mx-auto mb-4 opacity-30" />
                                    <p className="font-medium">Your cart is empty</p>
                                    <p className="text-sm mt-1">Add some products to get started</p>
                                </div>
                            ) : (
                                items.map((item) => (
                                    <motion.div
                                        key={item.id}
                                        layout
                                        initial={{ opacity: 0, x: 20 }}
                                        animate={{ opacity: 1, x: 0 }}
                                        exit={{ opacity: 0, x: 20 }}
                                        className="flex items-center gap-4 p-4 rounded-2xl bg-white/5 border border-white/8"
                                    >
                                        <div className="w-14 h-14 rounded-xl bg-gradient-to-br from-blue-500/20 to-purple-500/20 flex items-center justify-center flex-shrink-0 text-xl">
                                            🛍️
                                        </div>
                                        <div className="flex-1 min-w-0">
                                            <p className="font-semibold text-sm truncate">
                                                {item.product?.name || `Product ${item.product_id.slice(0, 8)}`}
                                            </p>
                                            <p className="text-xs text-gray-400 mt-0.5">{item.product?.category || ''}</p>
                                            <p className="text-xs text-blue-400 mt-1">Qty: {item.quantity} × ${item.unit_price.toFixed(2)}</p>
                                        </div>
                                        <div className="flex flex-col items-end gap-2 flex-shrink-0">
                                            <p className="font-bold text-sm">${(item.quantity * item.unit_price).toFixed(2)}</p>
                                            <button
                                                onClick={() => handleRemove(item.product_id)}
                                                disabled={removing[item.product_id]}
                                                className="p-1.5 rounded-lg text-gray-500 hover:text-red-400 hover:bg-red-400/10 transition-all disabled:opacity-40"
                                                title="Remove item"
                                            >
                                                {removing[item.product_id]
                                                    ? <div className="w-4 h-4 border-2 border-red-400/30 border-t-red-400 rounded-full animate-spin" />
                                                    : <Trash2 className="w-4 h-4" />}
                                            </button>
                                        </div>
                                    </motion.div>
                                ))
                            )}
                        </div>

                        {/* Footer */}
                        {!orderComplete && items.length > 0 && (
                            <div className="px-6 py-5 border-t border-white/10 space-y-4">
                                <div className="flex justify-between items-center">
                                    <span className="text-gray-400">Subtotal</span>
                                    <span className="text-2xl font-black">${total.toFixed(2)}</span>
                                </div>
                                <button
                                    onClick={handleCheckout}
                                    disabled={checkingOut}
                                    className="w-full py-4 rounded-2xl bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-500 hover:to-purple-500 text-white font-bold flex items-center justify-center gap-2 transition-all disabled:opacity-50 disabled:cursor-not-allowed shadow-lg"
                                >
                                    {checkingOut ? (
                                        <div className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                                    ) : (
                                        <>
                                            <CreditCard className="w-5 h-5" />
                                            Checkout
                                        </>
                                    )}
                                </button>
                            </div>
                        )}
                    </motion.div>
                </>
            )}
        </AnimatePresence>
    );
}
