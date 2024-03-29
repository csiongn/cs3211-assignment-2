import random

# Configuration Parameters
num_threads = 40
orders = 100000
num_instruments = 5
orders_gen=list()
instruments = [f"SYM{i}" for i in range(num_instruments)]
order_id = 0

# Generate commands for a single thread
def generate_commands():
    global order_id
    for _ in range(orders):
        type = random.choice(['B', 'S', 'C'])
        instrument = random.choice(instruments)
        thread_id = random.randint(0, num_threads - 1)
        price, count = random.randint(100, 10000), random.randint(1, 100)
        if type == 'C':
            cancelled_order = random.randint(0, order_id - 1)
            print(f"{orders_gen[cancelled_order][0]} C {orders_gen[cancelled_order][1]}")
        else:
            orders_gen.append((thread_id, order_id))
            print(f"{thread_id} {type} {order_id} {instrument} {price} {count}")
            order_id += 1

def main():
    print(num_threads)

    # Open all threads
    print("o")

    # Barrier to sync all threads
    print(".")

    # Generate and print commands
    generate_commands()

    print("x")

if __name__ == "__main__":
    main()
