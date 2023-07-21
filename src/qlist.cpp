#include <iostream>
#include <wx/wxprec.h>
#include <wx/filedlg.h>
#include <wx/dataview.h>
#ifndef WX_PRECOMP
#include <wx/wx.h>
#endif

using namespace std;
class Node;
using NodePtr = unique_ptr<Node>;
using NodePtrArray = vector<NodePtr>;

class Node
{
public:
    Node(Node *parent,
         const wxString &key, const wxString &type,
         const wxString &value)
    {
        m_parent = parent;
        m_key = key;
        m_type = type;
        m_value = value;
    }

    ~Node() = default;

    bool IsContainer() const
    {
        return m_children.size() > 0;
    }

    Node *GetParent()
    {
        return m_parent;
    }
    NodePtrArray &GetChildren()
    {
        return m_children;
    }
    Node *GetNthChild(unsigned int n)
    {
        return m_children.at(n).get();
    }
    void Insert(Node *child, unsigned int n)
    {
        m_children.insert(m_children.begin() + n, NodePtr(child));
    }
    Node *AddEntry(const wxString &key, const wxString &type,
                   const wxString &value)
    {
        Node *node = new Node(this, key, type, value);
        m_children.push_back(NodePtr(node));
        return node;
    }
    void Append(Node *child)
    {
        m_children.push_back(NodePtr(child));
    }
    unsigned int GetChildCount() const
    {
        return m_children.size();
    }

public:
    wxString m_key;
    wxString m_type;
    wxString m_value;

private:
    Node *m_parent;
    NodePtrArray m_children;
};

class Model : public wxDataViewModel
{
public:
    Model();
    ~Model()
    {
        delete m_root;
    }
    virtual void GetValue(wxVariant &variant,
                          const wxDataViewItem &item, unsigned int col) const override;
    virtual bool SetValue(const wxVariant &variant,
                          const wxDataViewItem &item, unsigned int col) override;
    virtual wxDataViewItem GetParent(const wxDataViewItem &item) const override;
    virtual bool IsContainer(const wxDataViewItem &item) const override;
    virtual unsigned int GetChildren(const wxDataViewItem &parent,
                                     wxDataViewItemArray &array) const override;
    Node *AddRootEntry(const wxString &key, const wxString &type,
                       const wxString &value) const;

private:
    Node *m_root;
};

Model::Model()
{
    m_root = new Node(nullptr, "Root", "Dictionary", "12 children");
};
Node *Model::AddRootEntry(const wxString &key, const wxString &type,
                          const wxString &value) const
{
    Node *node = new Node(m_root, key, type, value);
    m_root->Append(node);
    return node;
};

void Model::GetValue(wxVariant &variant,
                     const wxDataViewItem &item, unsigned int col) const
{
    wxASSERT(item.IsOk());

    Node *node = (Node *)item.GetID();
    switch (col)
    {
    case 0:
        variant = node->m_key;
        break;
    case 1:
        variant = node->m_type;
        break;
    case 2:
        variant = node->m_value;
        break;
    default:
        wxLogError("Model::GetValue: wrong column %d", col);
    }
};

bool Model::SetValue(const wxVariant &variant,
                     const wxDataViewItem &item, unsigned int col)
{
    wxASSERT(item.IsOk());

    Node *node = (Node *)item.GetID();
    switch (col)
    {
    case 0:
        node->m_key = variant.GetString();
        return true;
    case 1:
        node->m_type = variant.GetString();
        return true;
    case 2:
        node->m_value = variant.GetString();
        return true;

    default:
        wxLogError("Model::SetValue: wrong column");
    }
    return false;
}

wxDataViewItem Model::GetParent(const wxDataViewItem &item) const
{
    if (!item.IsOk())
        return wxDataViewItem(0);

    Node *node = (Node *)item.GetID();

    if (node == m_root)
        return wxDataViewItem(0);

    return wxDataViewItem((void *)node->GetParent());
}

bool Model::IsContainer(const wxDataViewItem &item) const
{
    if (!item.IsOk())
        return true;

    Node *node = (Node *)item.GetID();
    return node->IsContainer();
}

unsigned int Model::GetChildren(const wxDataViewItem &parent,
                                wxDataViewItemArray &array) const
{
    Node *node = (Node *)parent.GetID();
    if (!node)
    {
        array.Add(wxDataViewItem((void *)m_root));
        return 1;
    }

    if (node->GetChildCount() == 0)
    {
        return 0;
    }

    for (const auto &child : node->GetChildren())
    {
        array.Add(wxDataViewItem(child.get()));
    }

    return array.size();
}

class MyApp : public wxApp
{
public:
    virtual bool OnInit();
};
class Frame : public wxFrame
{
public:
    Frame(const wxString &title, const wxPoint &pos, const wxSize &size);

private:
    void OnFileOpen(wxCommandEvent &event);
    void OnExit(wxCommandEvent &event);
    void OnAbout(wxCommandEvent &event);
    wxDataViewCtrl *tree;
    wxDECLARE_EVENT_TABLE();
};
enum
{
    ID_FILE = 1,
    ID_NEW = 2
};
wxBEGIN_EVENT_TABLE(Frame, wxFrame)
    EVT_MENU(ID_FILE, Frame::OnFileOpen)
        EVT_MENU(wxID_EXIT, Frame::OnExit)
            EVT_MENU(wxID_ABOUT, Frame::OnAbout)
                wxEND_EVENT_TABLE()
                    wxIMPLEMENT_APP(MyApp);
bool MyApp::OnInit()
{
    Frame *frame = new Frame("Qlist", wxPoint(50, 50), wxSize(550, 450));
    frame->Show(true);
    return true;
}
Frame::Frame(const wxString &title, const wxPoint &pos, const wxSize &size)
    : wxFrame(NULL, wxID_ANY, title, pos, size)
{
    wxMenu *menuFile = new wxMenu;
    menuFile->Append(ID_FILE, "&Open\tCtrl-O");
    menuFile->Append(ID_NEW, "&New\tCtrl-N");
    menuFile->Append(wxID_EXIT);
    menuFile->Append(wxID_ABOUT, "&About Qlist");
    menuFile->Append(wxID_PREFERENCES, "&Settings");
    wxMenuBar *menuBar = new wxMenuBar;
    menuBar->Append(menuFile, "&File");
    SetMenuBar(menuBar);
    wxDataViewCtrl *dataview = new wxDataViewCtrl(this, wxID_ANY, wxDefaultPosition, wxDefaultSize, wxDV_ROW_LINES);
    dataview->AppendTextColumn("Key", 0);
    dataview->AppendTextColumn("Type", 1);
    dataview->AppendTextColumn("Value", 2);
    Model *model = new Model;
    model->AddRootEntry("hi", "hi1", "hi2"); //->AddEntry("hi2", "hi3", "hi4");
    dataview->AssociateModel(model);
}
void Frame::OnExit(wxCommandEvent &event)
{
    Close(true);
}
void Frame::OnAbout(wxCommandEvent &event)
{
    wxMessageBox("This is a wxWidgets' Hello world sample",
                 "About Hello World", wxOK | wxICON_INFORMATION);
}
void Frame::OnFileOpen(wxCommandEvent &event)
{
    wxFileDialog
        openFileDialog(this, _("Open Property-List file"), wxEmptyString, wxEmptyString,
                       _("Property-List file|*.plist"), wxFD_OPEN | wxFD_FILE_MUST_EXIST);
    if (openFileDialog.ShowModal() == wxID_OK)
        cout << "hey!";
}