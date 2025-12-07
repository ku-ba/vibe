from django.test import TestCase
from django.urls import reverse
from .models import Todo
from django.utils import timezone

class TodoModelTest(TestCase):
    def test_string_representation(self):
        todo = Todo(title="My Test Todo")
        self.assertEqual(str(todo), "My Test Todo")

class TodoViewTest(TestCase):
    def setUp(self):
        # Setup a seed TODO list with at least one TODO
        self.todo = Todo.objects.create(
            title="Seed Todo",
            description="Seed Description",
            due_date=timezone.now()
        )

    def test_todo_list_view(self):
        # Action: Get the list view
        response = self.client.get(reverse('todo_list'))
        
        # Check return code is 200
        self.assertEqual(response.status_code, 200)
        self.assertContains(response, "Seed Todo")
        self.assertTemplateUsed(response, 'home.html')

    def test_todo_create_view_get(self):
        # Action: Get the create page
        response = self.client.get(reverse('todo_create'))
        
        # Check return code is 200
        self.assertEqual(response.status_code, 200)
        self.assertTemplateUsed(response, 'todos/todo_form.html')

    def test_todo_create_view_post(self):
        # Action: Create a new todo
        response = self.client.post(reverse('todo_create'), {
            'title': 'New Todo',
            'description': 'New Description',
            'due_date': '2023-12-31'
        }, follow=True) # Follow redirect to get 200
        
        # Check return code is 200
        self.assertEqual(response.status_code, 200)
        
        # Assert state: New object created
        self.assertEqual(Todo.objects.count(), 2)
        self.assertTrue(Todo.objects.filter(title='New Todo').exists())

    def test_todo_update_view_get(self):
        # Action: Get the update page for seeded object
        response = self.client.get(reverse('todo_update', args=[self.todo.pk]))
        
        # Check return code is 200
        self.assertEqual(response.status_code, 200)
        self.assertTemplateUsed(response, 'todos/todo_form.html')

    def test_todo_update_view_post(self):
        # Action: Update the seeded object
        response = self.client.post(reverse('todo_update', args=[self.todo.pk]), {
            'title': 'Updated Todo',
            'description': 'Updated Description',
            'due_date': '2023-12-31'
        }, follow=True)
        
        # Check return code is 200
        self.assertEqual(response.status_code, 200)
        
        # Assert state: Seeded object updated
        self.todo.refresh_from_db()
        self.assertEqual(self.todo.title, 'Updated Todo')

    def test_todo_delete_view_get(self):
        # Action: Get the delete confirmation page
        response = self.client.get(reverse('todo_delete', args=[self.todo.pk]))
        
        # Check return code is 200
        self.assertEqual(response.status_code, 200)
        self.assertTemplateUsed(response, 'todos/todo_confirm_delete.html')

    def test_todo_delete_view_post(self):
        # Action: Delete the seeded object
        response = self.client.post(reverse('todo_delete', args=[self.todo.pk]), follow=True)
        
        # Check return code is 200
        self.assertEqual(response.status_code, 200)
        
        # Assert state: Seeded object deleted
        self.assertEqual(Todo.objects.count(), 0)

    def test_todo_toggle_view(self):
        # Initial state check
        self.assertFalse(self.todo.is_resolved)
        
        # Action: Toggle the seeded object
        response = self.client.get(reverse('todo_toggle', args=[self.todo.pk]), follow=True)
        
        # Check return code is 200
        self.assertEqual(response.status_code, 200)
        
        # Assert state: Seeded object toggled
        self.todo.refresh_from_db()
        self.assertTrue(self.todo.is_resolved)
